package config

// // NOTE: ONLY PUT STRUCT DEFINITIONS IN THIS FILE
// //
// // After making edits here, run 'make cfgdoc-gen' (or 'make gen')

// Common is common config between full node and miner
type Common struct {
	API API
}

// API contains configs for API endpoint
type API struct {
	ListenAddress       string
	RemoteListenAddress string
	Timeout             Duration
}

// ManagerCfg manager config
type ManagerCfg struct {
	Common
	// database address
	DatabaseAddress string
}

// ProviderCfg provider config
type ProviderCfg struct {
	Common
	// used when 'ListenAddress' is unspecified. must be a valid duration recognized by golang's time.ParseDuration function
	Timeout  string
	Owner    string
	HostURI  string
	PublicIP string

	KubeConfigPath string
}
