package domain

type Model struct {
	ID           string
	Library      string
	PipelineTag  string
	Parameters   uint64
	Architecture string
	ModelType    string

	HiddenSize            int
	HiddenLayers          int
	AttentionHeads        int
	KeyValueHeads         int
	MaxPositionEmbeddings int
	HeadDim               int

	Files []string
}

type Precision string

const (
	PrecisionFP32 Precision = "FP32"
	PrecisionFP16 Precision = "FP16"
	PrecisionINT8 Precision = "INT8"
	PrecisionINT4 Precision = "INT4"
)

type CompatibilityLevel string

const (
	LevelExcellent  CompatibilityLevel = "excellent"
	LevelGood       CompatibilityLevel = "good"
	LevelLimited    CompatibilityLevel = "limited"
	LevelPoor       CompatibilityLevel = "poor"
	LevelImpossible CompatibilityLevel = "impossible"
)

type CompatibilityResult struct {
	Precision Precision

	WeightsBytes  uint64
	KVCacheBytes  uint64
	OverheadBytes uint64
	RequiredVRAM  uint64

	AvailableVRAM uint64
	AvailableRAM  uint64

	ContextSize int

	Score int
	Level CompatibilityLevel

	CanFitVRAM        bool
	CanFitWithOffload bool

	Message string
}

func (r CompatibilityResult) RequiredVRAMGB() float64 {
	return BytesToGB(r.RequiredVRAM)
}

func (r CompatibilityResult) WeightsGB() float64 {
	return BytesToGB(r.WeightsBytes)
}

func (r CompatibilityResult) KVCacheGB() float64 {
	return BytesToGB(r.KVCacheBytes)
}

func (r CompatibilityResult) OverheadGB() float64 {
	return BytesToGB(r.OverheadBytes)
}

func BytesToGB(value uint64) float64 {
	return float64(value) / (1024 * 1024 * 1024)
}
