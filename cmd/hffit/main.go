package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alexperezortuno/hffit/internal/compatibility"
	"github.com/alexperezortuno/hffit/internal/domain"
	"github.com/alexperezortuno/hffit/internal/hardware"
	"github.com/alexperezortuno/hffit/internal/huggingface"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	modelID := os.Args[1]

	fmt.Printf("HF Fit %s\n\n", version)

	detector := hardware.NewDetector()

	hw, err := detector.Detect()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"hardware detection failed: %v\n",
			err,
		)

		os.Exit(1)
	}

	printHardware(hw)

	client := huggingface.NewClient()

	model, err := client.GetModel(
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
		fmt.Fprintf(
			os.Stderr,
			"model does not expose parameter metadata\n",
		)

		os.Exit(1)
	}

	printModel(model)

	calculator := compatibility.NewCalculator()

	results := calculator.Calculate(
		model,
		hw,
	)

	printCompatibility(results)
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("  hffit <huggingface-model>")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println()
	fmt.Println("  hffit Qwen/Qwen3-8B")
}

func printHardware(hw *domain.Hardware) {
	fmt.Println("Hardware")
	fmt.Println("──────────────────────────────────")

	fmt.Printf(
		"OS       %s/%s\n",
		hw.OS,
		hw.Arch,
	)

	fmt.Printf(
		"CPU      %s\n",
		hw.CPU.Model,
	)

	fmt.Printf(
		"Cores    %d\n",
		hw.CPU.Cores,
	)

	fmt.Printf(
		"RAM      %.2f GB\n",
		hw.Memory.TotalGB(),
	)

	if len(hw.GPUs) == 0 {
		fmt.Println("GPU      Not detected")
	}

	for _, gpu := range hw.GPUs {
		fmt.Printf(
			"GPU      %s\n",
			gpu.Model,
		)

		fmt.Printf(
			"VRAM     %.2f GB\n",
			gpu.VRAMGB(),
		)

		fmt.Printf(
			"Driver   %s\n",
			gpu.DriverVersion,
		)
	}

	fmt.Println()
}

func printModel(model *domain.Model) {
	fmt.Println("Model")
	fmt.Println("──────────────────────────────────")

	fmt.Printf(
		"ID       %s\n",
		model.ID,
	)

	fmt.Printf(
		"Type     %s\n",
		model.ModelType,
	)

	fmt.Printf(
		"Arch     %s\n",
		model.Architecture,
	)

	fmt.Printf(
		"Params   %.2f B\n",
		float64(model.Parameters)/1_000_000_000,
	)

	fmt.Println()
}

func printCompatibility(
	results []domain.CompatibilityResult,
) {

	fmt.Println("Compatibility")
	fmt.Println("──────────────────────────────────")

	for _, result := range results {
		fmt.Printf(
			"%s %s %-6s %3d/100  %6.2f GB  %s\n",
			icon(result.Level),
			levelName(result.Level),
			result.Precision,
			result.Score,
			result.RequiredVRAMGB(),
			result.Message,
		)
	}
}

func icon(
	level domain.CompatibilityLevel,
) string {

	switch level {

	case domain.LevelExcellent:
		return "🟢"

	case domain.LevelGood:
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
		return "Good     "

	case domain.LevelLimited:
		return "Limited  "

	case domain.LevelPoor:
		return "Poor     "

	default:
		return "No       "
	}
}
