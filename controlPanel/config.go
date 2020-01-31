package controlpanel

type Config struct {
	// web hosting config
	Host       string
	Port       int
	TLSEnabled bool
	KeyFile    string
	CertFile   string

	// page information config
	NodeName   string
	Version    string
	BuildNumer string

	// Display state information
	CompleteHeight uint32
	LeaderHeight   uint32
}
