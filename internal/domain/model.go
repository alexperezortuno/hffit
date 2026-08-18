package domain

type Model struct {
	ID           string
	Library      string
	PipelineTag  string
	Parameters   uint64
	Architecture string
	ModelType    string
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
	Precision         Precision
	RequiredVRAM      uint64
	AvailableVRAM     uint64
	AvailableRAM      uint64
	Score             int
	Level             CompatibilityLevel
	CanFitVRAM        bool
	CanFitWithOffload bool
	Message           string
}

func (r CompatibilityResult) RequiredVRAMGB() float64 {
	return float64(r.RequiredVRAM) / (1024 * 1024 * 1024)
}
