package controlpanel

type Config struct {
	// web hosting config
	Host       string
	Port       int
	TLSEnabled bool
	KeyFile    string
	CertFile   string

	// page information config
	FactomNodeName string
	Version        string
	BuildNumer     string
}
