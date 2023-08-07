package manifest

type ResourceUnits struct {
	CPU       *CPU
	Memory    *Memory
	Storage   []*Storage
	GPU       *GPU
	Endpoints []*Endpoint
}

func NewResourceUnits(cpu, memory, storage uint64) *ResourceUnits {
	return &ResourceUnits{CPU: NewCPU(cpu), Memory: NewMemory(memory), Storage: []*Storage{NewStorage(storage)}}
}
