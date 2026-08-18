package hardware

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/alexperezortuno/hffit/internal/domain"
)

type Detector struct{}

func NewDetector() *Detector {
	return &Detector{}
}

func (d *Detector) Detect() (*domain.Hardware, error) {
	hw := &domain.Hardware{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	cpu, err := detectCPU()
	if err == nil {
		hw.CPU = cpu
	}

	memory, err := detectMemory()
	if err != nil {
		return nil, fmt.Errorf("detect memory: %w", err)
	}

	hw.Memory = memory

	gpus, err := detectNvidiaGPUs()
	if err == nil {
		hw.GPUs = gpus
	}

	return hw, nil
}

func detectCPU() (domain.CPU, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return domain.CPU{}, err
	}
	defer file.Close()

	var cpu domain.CPU

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "model name") && cpu.Model == "" {
			parts := strings.SplitN(line, ":", 2)

			if len(parts) == 2 {
				cpu.Model = strings.TrimSpace(parts[1])
			}
		}

		if strings.HasPrefix(line, "processor") {
			cpu.Cores++
		}
	}

	return cpu, scanner.Err()
}

func detectMemory() (domain.Memory, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return domain.Memory{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "MemTotal:") {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) < 2 {
			return domain.Memory{}, fmt.Errorf("invalid MemTotal format")
		}

		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return domain.Memory{}, err
		}

		return domain.Memory{
			TotalBytes: kb * 1024,
		}, nil
	}

	return domain.Memory{}, fmt.Errorf("MemTotal not found")
}

func detectNvidiaGPUs() ([]domain.GPU, error) {
	cmd := exec.Command(
		"nvidia-smi",
		"--query-gpu=name,memory.total,driver_version",
		"--format=csv,noheader,nounits",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var gpus []domain.GPU

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		parts := strings.Split(line, ",")

		if len(parts) != 3 {
			continue
		}

		model := strings.TrimSpace(parts[0])
		memoryMB := strings.TrimSpace(parts[1])
		driver := strings.TrimSpace(parts[2])

		mb, err := strconv.ParseUint(memoryMB, 10, 64)
		if err != nil {
			continue
		}

		gpus = append(gpus, domain.GPU{
			Vendor:        "NVIDIA",
			Model:         model,
			VRAMBytes:     mb * 1024 * 1024,
			DriverVersion: driver,
		})
	}

	return gpus, nil
}
