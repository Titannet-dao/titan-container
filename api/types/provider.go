package types

import "time"

type ProviderID string

type ProviderState int

const (
	ProviderStateOnline ProviderState = iota + 1
	ProviderStateOffline
	ProviderStateAbnormal
)

func ProviderStateString(state ProviderState) string {
	switch state {
	case ProviderStateOnline:
		return "Online"
	case ProviderStateOffline:
		return "Offline"
	case ProviderStateAbnormal:
		return "Abnormal"
	default:
		return "Unknown"
	}
}

type Provider struct {
	ID        ProviderID    `db:"id"`
	Owner     string        `db:"owner"`
	HostURI   string        `db:"host_uri"`
	IP        string        `db:"ip"`
	State     ProviderState `db:"state"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}

type GetProviderOption struct {
	Owner string
	ID    ProviderID
	State []ProviderState
	Page  int
	Size  int
}

type ResourcesStatistics struct {
	Memory   Memory
	CPUCores CPUCores
	Storage  Storage
}

type Memory struct {
	MaxMemory uint64
	Available uint64
	Active    uint64
	Pending   uint64
}

type CPUCores struct {
	MaxCPUCores float64
	Available   float64
	Active      float64
	Pending     float64
}

type Storage struct {
	MaxStorage uint64
	Available  uint64
	Active     uint64
	Pending    uint64
}
