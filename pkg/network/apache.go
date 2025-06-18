package network

// ApacheConfigServer represents the Apache server that hosts network configuration files
type ApacheConfigServer interface {
	URL() string
	GenesisSSZURL() string
	ConfigYAMLURL() string
	BootnodesYAMLURL() string
	DepositContractBlockURL() string
}

// apacheConfigServer is the concrete implementation
type apacheConfigServer struct {
	url string
}

// NewApacheConfigServer creates a new Apache config server instance
func NewApacheConfigServer(url string) ApacheConfigServer {
	return &apacheConfigServer{url: url}
}

func (a *apacheConfigServer) URL() string {
	return a.url
}

func (a *apacheConfigServer) GenesisSSZURL() string {
	return a.url + "/network-configs/genesis.ssz"
}

func (a *apacheConfigServer) ConfigYAMLURL() string {
	return a.url + "/network-configs/config.yaml"
}

func (a *apacheConfigServer) BootnodesYAMLURL() string {
	return a.url + "/network-configs/boot_enr.yaml"
}

func (a *apacheConfigServer) DepositContractBlockURL() string {
	return a.url + "/network-configs/deposit_contract_block.txt"
}