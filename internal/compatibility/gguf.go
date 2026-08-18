package compatibility

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/alexperezortuno/hffit/internal/domain"
)

var quantizationPattern = regexp.MustCompile(
	`(?i)(IQ[1-4]_[A-Z0-9_]+|Q[2-8]_[A-Z0-9_]+|F16|F32|BF16)`,
)

var quantizationBits = map[string]float64{
	"F32":  32.0,
	"F16":  16.0,
	"BF16": 16.0,

	"Q8_0": 8.5,

	"Q6_K": 6.6,

	"Q5_K_M": 5.7,
	"Q5_K_S": 5.5,

	"Q4_K_M": 4.8,
	"Q4_K_S": 4.6,
	"Q4_0":   4.5,

	"Q3_K_L": 3.8,
	"Q3_K_M": 3.6,
	"Q3_K_S": 3.4,

	"Q2_K": 2.8,
}

func DiscoverGGUF(
	model *domain.Model,
) []domain.GGUFVariant {

	var variants []domain.GGUFVariant

	for _, filename := range model.Files {

		if !strings.EqualFold(
			filepath.Ext(filename),
			".gguf",
		) {
			continue
		}

		quant := detectQuantization(filename)

		if quant == "" {
			continue
		}

		bits := bitsPerWeight(quant)

		variants = append(
			variants,
			domain.GGUFVariant{
				Filename:      filename,
				Quantization:  quant,
				BitsPerWeight: bits,
			},
		)
	}

	sort.Slice(
		variants,
		func(i, j int) bool {
			return variants[i].BitsPerWeight >
				variants[j].BitsPerWeight
		},
	)

	return variants
}

func detectQuantization(
	filename string,
) string {

	filename =
		strings.ToUpper(filename)

	match :=
		quantizationPattern.FindString(filename)

	if match == "" {
		return ""
	}

	return strings.ToUpper(match)
}

func bitsPerWeight(
	quantization string,
) float64 {

	if bits, ok :=
		quantizationBits[quantization]; ok {

		return bits
	}

	return fallbackQuantBits(
		quantization,
	)
}

func fallbackQuantBits(
	quant string,
) float64 {

	switch {

	case strings.HasPrefix(quant, "Q8"):
		return 8.5

	case strings.HasPrefix(quant, "Q6"):
		return 6.6

	case strings.HasPrefix(quant, "Q5"):
		return 5.7

	case strings.HasPrefix(quant, "Q4"):
		return 4.8

	case strings.HasPrefix(quant, "Q3"):
		return 3.6

	case strings.HasPrefix(quant, "Q2"):
		return 2.8

	default:
		return 0
	}
}
