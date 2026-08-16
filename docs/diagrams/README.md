# Architecture Diagrams

This directory contains Mermaid diagrams for visualizing the Jitter Distill architecture and data flow.

## Viewing Diagrams

### In GitHub

All `.md` files with Mermaid code blocks are automatically rendered in GitHub's markdown viewer.

### In VS Code

Install the [Markdown Preview Mermaid Support](https://marketplace.visualstudio.com/items?itemName=bierner.markdown-mermaid) extension.

### In Documentation

Use [Mermaid Live Editor](https://mermaid.live/) or tools like:
- `mmdc` (Mermaid CLI)
- MkDocs with mermaid plugin
- GitHub Pages with mermaid support

## Files

- `flow.md` - Architecture diagrams and data flows
- `animated-flow.md` - Dynamic pipeline visualization

## Rendering Commands

```bash
# Install Mermaid CLI
npm install -g @mermaid-js/mermaid-cli

# Render diagrams
mmdc -i docs/diagrams/flow.md -o docs/diagrams/flow.svg
mmdc -i docs/diagrams/animated-flow.md -o docs/diagrams/animated-flow.svg
```
