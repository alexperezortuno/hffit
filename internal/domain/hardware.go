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
	return BytesToGB(m.TotalBytes)
}

func (g GPU) VRAMGB() float64 {
	return BytesToGB(g.VRAMBytes)
}
