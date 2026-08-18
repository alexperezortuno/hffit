package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/alexperezortuno/hffit/internal/compatibility"
	"github.com/alexperezortuno/hffit/internal/domain"
	"github.com/alexperezortuno/hffit/internal/hardware"
	"github.com/alexperezortuno/hffit/internal/huggingface"
)

const version = "0.3.2"

func main() {

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {

	case "--version", "-v":
		fmt.Printf(
			"hffit %s\n",
			version,
		)
		return

	case "--help", "-h":
		printUsage()
		return
	}

	modelID := os.Args[1]

	flags := flag.NewFlagSet(
		"hffit",
		flag.ExitOnError,
	)

	contextSize := flags.Int(
		"context",
		8192,
		"context window in tokens",
	)

	_ = flags.Parse(
		os.Args[2:],
	)

	if *contextSize <= 0 {
		fmt.Fprintln(
			os.Stderr,
			"context must be greater than zero",
		)

		os.Exit(1)
	}

	fmt.Printf(
		"HF Fit %s\n\n",
		version,
	)

	detector :=
		hardware.NewDetector()

	hw, err :=
		detector.Detect()

	if err != nil {

		fmt.Fprintf(
			os.Stderr,
			"hardware detection failed: %v\n",
			err,
		)

		os.Exit(1)
	}

	client :=
		huggingface.NewClient()

	model, err :=
		client.GetModel(
			context.Background(),
			modelID,
		)

	if err != nil {

		fmt.Fprintf(
			os.Stderr,
			"model lookup failed: %v\n",
			err,
		)

		os.Exit(1)
	}

	if model.Parameters == 0 {

		fmt.Fprintln(
			os.Stderr,
			"model does not expose parameter metadata",
		)

		os.Exit(1)
	}

	if model.MaxPositionEmbeddings > 0 &&
		*contextSize > model.MaxPositionEmbeddings {

		fmt.Fprintf(
			os.Stderr,
			"warning: requested context %d exceeds model configured context %d\n\n",
			*contextSize,
			model.MaxPositionEmbeddings,
		)
	}

	printHardware(hw)

	printModel(
		model,
		*contextSize,
	)

	calculator :=
		compatibility.NewCalculator(
			compatibility.Options{
				ContextSize: *contextSize,
			},
		)

	results :=
		calculator.Calculate(
			model,
			hw,
		)

	printCompatibility(results)

	ggufVariants :=
		compatibility.DiscoverGGUF(model)

	if len(ggufVariants) > 0 {

		ggufCalculator :=
			compatibility.NewGGUFCalculator(
				*contextSize,
			)

		ggufResults :=
			ggufCalculator.Calculate(
				model,
				hw,
				ggufVariants,
			)

		printGGUFCompatibility(
			model,
			ggufResults,
			*contextSize,
		)
	}
}

func printUsage() {

	fmt.Println(`
HF Fit - Hugging Face model compatibility checker

Usage:

  hffit <model> [options]

Examples:

  hffit Qwen/Qwen3-8B

  hffit Qwen/Qwen3-8B --context 32768

  hffit https://huggingface.co/Qwen/Qwen3-8B --context 131072

Options:

  --context <tokens>
        Context window to evaluate.
        Default: 8192

  --version
        Show version
`)
}

func printHardware(
	hw *domain.Hardware,
) {

	fmt.Println("Hardware")
	fmt.Println("────────────────────────────────────────")

	fmt.Printf(
		"OS          %s/%s\n",
		hw.OS,
		hw.Arch,
	)

	fmt.Printf(
		"CPU         %s\n",
		hw.CPU.Model,
	)

	fmt.Printf(
		"Cores       %d\n",
		hw.CPU.Cores,
	)

	fmt.Printf(
		"RAM         %.2f GB\n",
		hw.Memory.TotalGB(),
	)

	if len(hw.GPUs) == 0 {

		fmt.Println(
			"GPU         Not detected",
		)
	}

	for _, gpu := range hw.GPUs {

		fmt.Printf(
			"GPU         %s\n",
			gpu.Model,
		)

		fmt.Printf(
			"VRAM        %.2f GB\n",
			gpu.VRAMGB(),
		)

		fmt.Printf(
			"Driver      %s\n",
			gpu.DriverVersion,
		)
	}

	fmt.Println()
}

func printModel(
	model *domain.Model,
	contextSize int,
) {

	fmt.Println("Model")
	fmt.Println("────────────────────────────────────────")

	fmt.Printf(
		"ID          %s\n",
		model.ID,
	)

	fmt.Printf(
		"ID          %s\n",
		model.ID,
	)

	if model.BaseModelID != "" {

		fmt.Printf(
			"Base model  %s\n",
			model.BaseModelID,
		)
	}

	fmt.Printf(
		"Type        %s\n",
		model.ModelType,
	)

	fmt.Printf(
		"Arch        %s\n",
		model.Architecture,
	)

	fmt.Printf(
		"Params      %.2f B\n",
		float64(model.Parameters)/
			1_000_000_000,
	)

	fmt.Printf(
		"Layers      %d\n",
		model.HiddenLayers,
	)

	fmt.Printf(
		"Hidden      %d\n",
		model.HiddenSize,
	)

	fmt.Printf(
		"Heads       %d\n",
		model.AttentionHeads,
	)

	fmt.Printf(
		"KV heads    %d\n",
		model.KeyValueHeads,
	)

	fmt.Printf(
		"Head dim    %d\n",
		model.HeadDim,
	)

	fmt.Printf(
		"Max context %d\n",
		model.MaxPositionEmbeddings,
	)

	fmt.Printf(
		"Test context %d\n",
		contextSize,
	)

	fmt.Println()
}

func printCompatibility(
	results []domain.CompatibilityResult,
) {

	fmt.Println("Compatibility")
	fmt.Println("──────────────────────────────────────────────────────────────────────────")

	for _, result := range results {

		fmt.Printf(
			"%s %-9s %-5s %3d/100  Total %6.2f GB  %s\n",
			icon(result.Level),
			levelName(result.Level),
			result.Precision,
			result.Score,
			result.RequiredVRAMGB(),
			result.Message,
		)

		fmt.Printf(
			"   Weights: %6.2f GB | KV: %6.2f GB | Overhead: %5.2f GB\n",
			result.WeightsGB(),
			result.KVCacheGB(),
			result.OverheadGB(),
		)
	}
}

func icon(
	level domain.CompatibilityLevel,
) string {

	switch level {

	case domain.LevelExcellent,
		domain.LevelGood:

		return "🟢"

	case domain.LevelLimited:

		return "🟡"

	case domain.LevelPoor:

		return "🟠"

	default:

		return "🔴"
	}
}

func levelName(
	level domain.CompatibilityLevel,
) string {

	switch level {

	case domain.LevelExcellent:
		return "Excellent"

	case domain.LevelGood:
		return "Good"

	case domain.LevelLimited:
		return "Limited"

	case domain.LevelPoor:
		return "Poor"

	default:
		return "No"
	}
}

func printGGUFCompatibility(
	model *domain.Model,
	results []domain.GGUFVariant,
	contextSize int,
) {

	fmt.Println()
	fmt.Println("GGUF")
	fmt.Println(
		"──────────────────────────────────────────────────────────────────────────",
	)

	for _, result := range results {

		recommended := ""

		if result.Recommended {
			recommended =
				" ← RECOMMENDED"
		}

		fmt.Printf(
			"%s %-8s %3d/100  %6.2f GB  %-35s%s\n",
			icon(result.Level),
			result.Quantization,
			result.Score,
			result.EstimatedVRAMGB(),
			result.Message,
			recommended,
		)

		fmt.Printf(
			"   Weights: %.2f GB | %.2f bpw | %s\n",
			result.EstimatedWeightsGB(),
			result.BitsPerWeight,
			result.Filename,
		)
	}

	fmt.Println()

	printRecommendedCommand(
		model,
		results,
		contextSize,
	)
}

func printRecommendedCommand(
	model *domain.Model,
	results []domain.GGUFVariant,
	contextSize int,
) {

	for _, result := range results {

		if !result.Recommended {
			continue
		}

		fmt.Println(
			"Recommended configuration",
		)

		fmt.Println(
			"────────────────────────────────────────",
		)

		fmt.Printf(
			"Quantization: %s\n",
			result.Quantization,
		)

		fmt.Printf(
			"Estimated VRAM: %.2f GB\n",
			result.EstimatedVRAMGB(),
		)

		fmt.Printf(
			"Context: %d\n\n",
			contextSize,
		)

		fmt.Println("llama.cpp:")

		fmt.Printf(
			"llama-cli -hf %s:%s -ngl 999 -c %d\n",
			model.ID,
			result.Quantization,
			contextSize,
		)

		fmt.Println()

		fmt.Println("Ollama:")

		fmt.Printf(
			"ollama run hf.co/%s:%s\n",
			model.ID,
			result.Quantization,
		)

		return
	}
}
