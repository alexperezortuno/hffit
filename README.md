# hffit

**HF Fit** - Check if a Hugging Face model fits on your hardware.

## Overview

`hffit` is a CLI tool that analyzes your hardware and a Hugging Face model to determine compatibility across different quantization levels. It outputs memory requirements, compatibility scores, and ready-to-run commands for llama.cpp and Ollama.

## Installation

```bash
go install github.com/alexperezortuno/hffit/cmd/hffit@latest
```

Or build from source:

```bash
git clone https://github.com/alexperezortuno/hffit
cd hffit
go build -o hffit ./cmd/hffit
```

## Usage

```bash
hffit <model-id> [options]
```

### Arguments

- `<model-id>` - Hugging Face model ID (e.g., `Qwen/Qwen3-8B`) or full HF URL

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--context` | `8192` | Context window in tokens |
| `--version` | - | Show version |
| `--help` | - | Show help |

### Examples

```bash
# Basic check
hffit Qwen/Qwen3-8B

# With custom context
hffit Qwen/Qwen3-8B --context 32768

# Using full HF URL
hffit https://huggingface.co/Qwen/Qwen3-8B --context 131072
```

## Output

The tool provides:

1. **Hardware** - Detected CPU, RAM, GPU(s), VRAM
2. **Model** - Architecture, parameters, layers, context limits
3. **Compatibility** - Scores for fp16, fp8, q8, q6, q5, q4, q3, q2
4. **GGUF** - Discovered GGUF variants with VRAM estimates
5. **Recommended** - Best quantization + ready-to-run commands

### Sample Output

```
HF Fit 0.3.2

Hardware
────────────────────────────────────────
OS          linux/amd64
CPU         AMD Ryzen 9 7950X
Cores       16
RAM         64.00 GB
GPU         NVIDIA RTX 4090
VRAM        24.00 GB
Driver      560.35.03

Model
────────────────────────────────────────
ID          Qwen/Qwen3-8B
Type        causal_lm
Arch        qwen3
Params      8.00 B
Layers      36
Hidden      4096
Heads       32
KV heads    8
Head dim    128
Max context 131072
Test context 8192

Compatibility
──────────────────────────────────────────────────────────────────────────
🟢 Excellent fp16       100/100  Total  16.00 GB  Fits in VRAM
   Weights:  16.00 GB | KV:  0.25 GB | Overhead:  1.00 GB
🟢 Excellent q8_0       100/100  Total   8.50 GB  Fits in VRAM
   Weights:   8.00 GB | KV:  0.25 GB | Overhead:  0.50 GB
...

GGUF
──────────────────────────────────────────────────────────────────────────
🟢 q4_k_m      95/100   5.50 GB  Recommended for most use cases ← RECOMMENDED
   Weights: 4.50 GB | 4.50 bpw | Qwen3-8B-Q4_K_M.gguf

Recommended configuration
────────────────────────────────────────
Quantization: q4_k_m
Estimated VRAM: 5.50 GB
Context: 8192

llama.cpp:
llama-cli -hf Qwen/Qwen3-8B:q4_k_m -ngl 999 -c 8192

Ollama:
ollama run hf.co/Qwen/Qwen3-8B:q4_k_m
```

## Compatibility Levels

| Level | Score | Meaning |
|-------|-------|---------|
| 🟢 Excellent | 90-100 | Fits comfortably with headroom |
| 🟢 Good | 75-89 | Fits with minimal overhead |
| 🟡 Limited | 50-74 | Fits but may need CPU offload |
| 🟠 Poor | 25-49 | Likely needs significant CPU offload |
| 🔴 No | 0-24 | Won't fit on this hardware |

## How It Works

1. **Hardware Detection** - Uses system APIs to detect CPU, RAM, and NVIDIA GPUs (via nvidia-smi)
2. **Model Fetching** - Queries Hugging Face API for model metadata (architecture, params, config)
3. **VRAM Calculation** - Computes weights + KV cache + overhead for each precision
4. **GGUF Discovery** - Finds available GGUF quantizations on HF Hub
5. **Scoring** - Rates each variant based on VRAM fit and quality

## Requirements

- Go 1.21+ (for building)
- NVIDIA GPU with nvidia-smi (for VRAM detection)
- Internet access (for Hugging Face API)

## License

MIT