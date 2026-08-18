package domain

type Hardware struct {
	OS     string
	Arch   string
	CPU    CPU
	Memory Memory
	GPUs   []GPU
}

type CPU struct {
	Model string
	Cores int
}

type Memory struct {
	TotalBytes uint64
}

type GPU struct {
	Vendor        string
	Model         string
	VRAMBytes     uint64
	DriverVersion string
}

func (m Memory) TotalGB() float64 {
	return bytesToGB(m.TotalBytes)
}

func (g GPU) VRAMGB() float64 {
	return bytesToGB(g.VRAMBytes)
}

func bytesToGB(value uint64) float64 {
	return float64(value) / (1024 * 1024 * 1024)
}
