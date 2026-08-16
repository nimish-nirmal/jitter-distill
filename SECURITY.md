# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, please email contact@nimishnirmal.dev

Please include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will acknowledge within 48 hours and provide a detailed response within 7 days.

## Security Considerations

### Entropy Quality

This tool distills hardware jitter but does not create entropy. Ensure:
- Sufficient physical entropy sources exist
- Pool has been running for adequate time
- Entropy estimation shows adequate bits (>128 bits recommended)

### Deployment

- Use in multi-worker mode for production
- Configure appropriate reseed intervals
- Never use predictable salts in production
- Consider hardware TRNG for critical applications

### Known Limitations

- VM/container environments reduce jitter quality
- Early boot systems may have limited entropy
- Not suitable as sole entropy source for high-security applications
