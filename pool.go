package jitterdistill

import (
	"bytes"
	"compress/gzip"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultSampleCount  = 512
	DefaultAROrder      = 8
	DefaultLearningRate = 0.0001
	DefaultWorkerCount  = 4
	DefaultChannelBuffer = 1024
	DefaultPoolSize     = 64
)

var ErrPoolShutdown = errors.New("entropy pool is shut down")

type EntropyPool struct {
	config         Config
	arPredictor    *ARPredictor
	sampleCh       chan []float64
	wg             sync.WaitGroup
	closed         int32
	shutdownOnce   sync.Once
	harvestDone    chan struct{}
	workersActive  int32
	bytesGenerated uint64
	tokensGenerated uint64
	lastReseed     time.Time
	reseedInterval time.Duration
	mu             sync.RWMutex
}

type Config struct {
	SampleCount  int
	AROrder      int
	LearningRate float64
	WorkerCount  int
	BufferSize   int
	PoolSize     int
	ReseedPeriod time.Duration
	Salt         []byte
}

func DefaultConfig() Config {
	return Config{
		SampleCount:  DefaultSampleCount,
		AROrder:      DefaultAROrder,
		LearningRate: DefaultLearningRate,
		WorkerCount:  DefaultWorkerCount,
		BufferSize:   DefaultChannelBuffer,
		PoolSize:     DefaultPoolSize,
		ReseedPeriod: 30 * time.Second,
		Salt:         []byte("ML-Jitter-Entropy-v1"),
	}
}

type ARPredictor struct {
	weights   []float64
	history   []float64
	order     int
	learnRate float64
}

func NewARPredictor(order int, lr float64) *ARPredictor {
	if order <= 0 {
		order = DefaultAROrder
	}
	if lr <= 0 || math.IsNaN(lr) || math.IsInf(lr, 0) {
		lr = DefaultLearningRate
	}
	return &ARPredictor{
		weights:   make([]float64, order),
		history:   make([]float64, order),
		order:     order,
		learnRate: lr,
	}
}

func (a *ARPredictor) Predict() float64 {
	var pred float64
	for i := 0; i < a.order; i++ {
		pred += a.weights[i] * a.history[i]
	}
	return pred
}

func (a *ARPredictor) Update(actual float64) float64 {
	if math.IsNaN(actual) || math.IsInf(actual, 0) {
		return 0.0
	}
	pred := a.Predict()
	err := actual - pred
	if math.IsNaN(err) || math.IsInf(err, 0) {
		err = 0.0
	}
	for i := 0; i < a.order; i++ {
		delta := a.learnRate * err * a.history[i]
		if !math.IsNaN(delta) && !math.IsInf(delta, 0) {
			a.weights[i] += delta
		}
		if math.IsNaN(a.weights[i]) || math.IsInf(a.weights[i], 0) {
			a.weights[i] = 0.0
		}
		if a.weights[i] > 1e6 {
			a.weights[i] = 1e6
		} else if a.weights[i] < -1e6 {
			a.weights[i] = -1e6
		}
	}
	for i := a.order - 1; i > 0; i-- {
		a.history[i] = a.history[i-1]
	}
	a.history[0] = actual
	return err
}

func (a *ARPredictor) Reset() {
	for i := range a.weights {
		a.weights[i] = 0.0
	}
	for i := range a.history {
		a.history[i] = 0.0
	}
}

func NewEntropyPool(cfg Config) *EntropyPool {
	if cfg.SampleCount <= 0 {
		cfg.SampleCount = DefaultSampleCount
	}
	if cfg.AROrder <= 0 {
		cfg.AROrder = DefaultAROrder
	}
	if cfg.LearningRate <= 0 {
		cfg.LearningRate = DefaultLearningRate
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = DefaultWorkerCount
	}
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = DefaultChannelBuffer
	}
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = DefaultPoolSize
	}
	if cfg.Salt == nil {
		cfg.Salt = []byte("ML-Jitter-Entropy-v1")
	}
	if cfg.ReseedPeriod <= 0 {
		cfg.ReseedPeriod = 30 * time.Second
	}
	p := &EntropyPool{
		config:         cfg,
		arPredictor:    NewARPredictor(cfg.AROrder, cfg.LearningRate),
		sampleCh:       make(chan []float64, cfg.BufferSize),
		harvestDone:    make(chan struct{}),
		reseedInterval: cfg.ReseedPeriod,
		lastReseed:     time.Now(),
	}
	p.startWorkers()
	p.startReseedMonitor()
	return p
}

func (p *EntropyPool) startWorkers() {
	p.wg.Add(p.config.WorkerCount)
	atomic.StoreInt32(&p.workersActive, int32(p.config.WorkerCount))
	for i := 0; i < p.config.WorkerCount; i++ {
		go p.harvestWorker()
	}
}

func (p *EntropyPool) harvestWorker() {
	defer p.wg.Done()
	defer atomic.AddInt32(&p.workersActive, -1)
	buf := make([]float64, p.config.SampleCount)
	for {
		select {
		case <-p.harvestDone:
			return
		default:
			samples := p.harvestJitter(buf)
			select {
			case p.sampleCh <- samples:
			case <-p.harvestDone:
				return
			}
		}
	}
}

func (p *EntropyPool) harvestJitter(buf []float64) []float64 {
	volatile := make([]byte, 4096)
	now := time.Now().UnixNano()
	samples := p.config.SampleCount
	if samples > len(buf) {
		samples = len(buf)
	}
	for i := 0; i < samples; i++ {
		for j := 0; j < len(volatile); j += 64 {
			volatile[j] ^= byte((i + j) & 0xFF)
			_ = volatile[j]
		}
		then := time.Now().UnixNano()
		delta := float64(then - now)
		now = then
		if math.IsNaN(delta) || math.IsInf(delta, 0) || delta < 0 {
			delta = 0.0
		}
		buf[i] = delta
	}
	return buf[:samples]
}

func (p *EntropyPool) startReseedMonitor() {
	go func() {
		ticker := time.NewTicker(p.reseedInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.Reseed()
			case <-p.harvestDone:
				return
			}
		}
	}()
}

func (p *EntropyPool) Reseed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.arPredictor.Reset()
	p.lastReseed = time.Now()
}

func (p *EntropyPool) GenerateToken() (string, error) {
	return p.GenerateTokenWithStrength(256)
}

func (p *EntropyPool) GenerateTokenWithStrength(bits int) (string, error) {
	if bits != 256 && bits != 512 {
		return "", fmt.Errorf("unsupported token strength: %d bits", bits)
	}
	if atomic.LoadInt32(&p.closed) != 0 {
		return "", ErrPoolShutdown
	}
	var residualBytes []byte
	byteBuf := make([]byte, 8)
	neededBytes := bits / 8
	samplesNeeded := neededBytes / 4
	if samplesNeeded < p.config.SampleCount {
		samplesNeeded = p.config.SampleCount
	}
	collected := 0
	for collected < samplesNeeded {
		select {
		case samples := <-p.sampleCh:
			for _, delta := range samples {
				residual := p.arPredictor.Update(delta)
				bitsVal := math.Float64bits(residual)
				binary.LittleEndian.PutUint64(byteBuf, bitsVal)
				residualBytes = append(residualBytes, byteBuf[:4]...)
				collected++
				if collected >= samplesNeeded {
					break
				}
			}
		case <-time.After(5 * time.Second):
			return "", errors.New("timeout waiting for entropy samples")
		}
	}
	hash := hmac.New(sha256.New, p.config.Salt)
	hash.Write(residualBytes)
	raw := hash.Sum(nil)
	if bits == 512 {
		hash2 := hmac.New(sha256.New, p.config.Salt)
		hash2.Write(raw)
		raw = append(raw, hash2.Sum(nil)...)
	}
	clear(residualBytes)
	atomic.AddUint64(&p.bytesGenerated, uint64(len(raw)))
	atomic.AddUint64(&p.tokensGenerated, 1)
	return hex.EncodeToString(raw), nil
}

func (p *EntropyPool) Stats() (bytesGenerated, tokensGenerated uint64, activeWorkers int32) {
	return atomic.LoadUint64(&p.bytesGenerated),
		atomic.LoadUint64(&p.tokensGenerated),
		atomic.LoadInt32(&p.workersActive)
}

func (p *EntropyPool) Close() error {
	p.shutdownOnce.Do(func() {
		atomic.StoreInt32(&p.closed, 1)
		close(p.harvestDone)
		p.wg.Wait()
		close(p.sampleCh)
	})
	return nil
}

func CompressToken(hexToken string) ([]byte, error) {
	if len(hexToken) == 0 {
		return nil, errors.New("empty token")
	}
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write([]byte(hexToken)); err != nil {
		w.Close()
		return nil, fmt.Errorf("compress write: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("compress close: %w", err)
	}
	return buf.Bytes(), nil
}

func DecompressToken(compressed []byte) (string, error) {
	if len(compressed) == 0 {
		return "", errors.New("empty compressed data")
	}
	r, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	defer r.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return "", fmt.Errorf("decompress read: %w", err)
	}
	return buf.String(), nil
}

func MixBytes(data []byte) {
	if len(data) == 0 {
		return
	}
	for i := 0; i < len(data)-7; i += 8 {
		val := float64(binary.LittleEndian.Uint64(data[i : i+8]))
		_ = val
	}
}

func EstimateEntropyBits(history []float64) float64 {
	if len(history) == 0 {
		return 0.0
	}
	var sum, sumSq float64
	n := float64(len(history))
	for _, v := range history {
		sum += v
		sumSq += v * v
	}
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	if variance < 0 {
		variance = 0
	}
	stdDev := math.Sqrt(variance)
	entropyBits := stdDev * 1e-3
	if entropyBits > 256 {
		entropyBits = 256
	}
	if entropyBits < 0 {
		entropyBits = 0
	}
	return entropyBits
}

func BitDiff(a, b []byte) int {
	if len(a) != len(b) {
		return -1
	}
	diff := 0
	for i := range a {
		diff += bits.OnesCount8(a[i] ^ b[i])
	}
	return diff
}

func VerifyTokenFormat(s string) bool {
	switch len(s) {
	case 64:
		return true
	case 128:
		return true
	default:
		return false
	}
}

// SignedToken represents a cryptographically signed token
type SignedToken struct {
	Token     string
	Signature []byte
	PublicKey []byte
}

// SignToken signs a token with RSA private key
func SignToken(token string, privateKeyPEM []byte) (*SignedToken, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	hash := sha256.Sum256([]byte(token))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	return &SignedToken{
		Token:     token,
		Signature: signature,
	}, nil
}

// VerifySignedToken verifies a signed token with RSA public key
func VerifySignedToken(signed *SignedToken, publicKeyPEM []byte) bool {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return false
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	rsaPub, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return false
	}

	hash := sha256.Sum256([]byte(signed.Token))
	err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hash[:], signed.Signature)
	return err == nil
}

// GenerateKeyPair generates RSA key pair for token signing
func GenerateKeyPair(bits int) (privateKeyPEM, publicKeyPEM []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyPEM = pem.EncodeToMemory(privateKeyBlock)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal public key: %w", err)
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyPEM = pem.EncodeToMemory(publicKeyBlock)

	return privateKeyPEM, publicKeyPEM, nil
}
