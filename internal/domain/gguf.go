package domain

type GGUFVariant struct {
	Filename      string
	Quantization  string
	BitsPerWeight float64

	EstimatedWeightsBytes uint64
	EstimatedVRAMBytes    uint64

	Score       int
	Level       CompatibilityLevel
	Recommended bool

	Message string
}

func (g GGUFVariant) EstimatedWeightsGB() float64 {
	return BytesToGB(g.EstimatedWeightsBytes)
}

func (g GGUFVariant) EstimatedVRAMGB() float64 {
	return BytesToGB(g.EstimatedVRAMBytes)
}
