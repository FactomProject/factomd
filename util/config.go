package util

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"time"

	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
	gcfg "gopkg.in/gcfg.v1"
)

var _ = fmt.Print

type FactomdConfig struct {
	App struct {
		PortNumber                             int
		HomeDir                                string
		ControlPanelPort                       int
		ControlPanelFilesPath                  string
		ControlPanelSetting                    string
		DBType                                 string
		LdbPath                                string
		BoltDBPath                             string
		DataStorePath                          string
		DirectoryBlockInSeconds                int
		ExportData                             bool
		ExportDataSubpath                      string
		FastBoot                               bool
		FastBootLocation                       string
		NodeMode                               string
		IdentityChainID                        string
		LocalServerPrivKey                     string
		LocalServerPublicKey                   string
		ExchangeRate                           uint64
		ExchangeRateChainId                    string
		ExchangeRateAuthorityPublicKey         string
		ExchangeRateAuthorityPublicKeyMainNet  string
		ExchangeRateAuthorityPublicKeyTestNet  string
		ExchangeRateAuthorityPublicKeyLocalNet string
		BitcoinAnchorRecordPublicKeys          []string
		EthereumAnchorRecordPublicKeys         []string

		// Network Configuration
		Network                 string
		MainNetworkPort         string
		PeersFile               string
		MainSeedURL             string
		MainSpecialPeers        string
		TestNetworkPort         string
		TestSeedURL             string
		TestSpecialPeers        string
		LocalNetworkPort        string
		LocalSeedURL            string
		LocalSpecialPeers       string
		CustomNetworkPort       string
		CustomSeedURL           string
		CustomSpecialPeers      string
		CustomBootstrapIdentity string
		CustomBootstrapKey      string
		P2PIncoming             int
		P2POutgoing             int
		FactomdTlsEnabled       bool
		FactomdTlsPrivateKey    string
		FactomdTlsPublicCert    string
		FactomdRpcUser          string
		FactomdRpcPass          string
		RequestTimeout          int
		RequestLimit            int
		CorsDomains             string

		ChangeAcksHeight uint32
	}
	Peer struct {
		AddPeers     []string      `short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
		ConnectPeers []string      `long:"connect" description:"Connect only to the specified peers at startup"`
		Listeners    []string      `long:"listen" description:"Add an interface/port to listen for connections (default all interfaces port: 8108, testnet: 18108)"`
		MaxPeers     int           `long:"maxpeers" description:"Max number of inbound and outbound peers"`
		BanDuration  time.Duration `long:"banduration" description:"How long to ban misbehaving peers.  Valid time units are {s, m, h}.  Minimum 1 second"`
		TestNet      bool          `long:"testnet" description:"Use the test network"`
		SimNet       bool          `long:"simnet" description:"Use the simulation test network"`
	}
	Log struct {
		LogPath         string
		LogLevel        string
		ConsoleLogLevel string
	}
	Wallet struct {
		Address          string
		Port             int
		DataFile         string
		RefreshInSeconds string
		BoltDBPath       string
		FactomdAddress   string
		FactomdPort      int
	}
	Walletd struct {
		WalletRpcUser       string
		WalletRpcPass       string
		WalletTlsEnabled    bool
		WalletTlsPrivateKey string
		WalletTlsPublicCert string
		FactomdLocation     string
		WalletdLocation     string
		WalletEncrypted     bool
	}
	LiveFeedAPI struct {
		EnableLiveFeedAPI     bool
		EventReceiverProtocol string
		EventReceiverAddress  string
		EventReceiverPort     int
		EventFormat           string
	}
}

// defaultConfig
const defaultConfig = `
; ------------------------------------------------------------------------------
; App settings
; ------------------------------------------------------------------------------
[app]
PortNumber                            = 8088
HomeDir                               = ""
; --------------- ControlPanel disabled | readonly | readwrite
ControlPanelSetting                   = readonly
ControlPanelPort                      = 8090
; --------------- DBType: LDB | Bolt | Map
DBType                                = "LDB"
LdbPath                               = "database/ldb"
BoltDBPath                            = "database/bolt"
DataStorePath                         = "data/export"
DirectoryBlockInSeconds               = 6
ExportData                            = false
ExportDataSubpath                     = "database/export/"
FastBoot                              = true
FastBootLocation                      = ""
; --------------- Network: MAIN | TEST | LOCAL
Network                               = MAIN
PeersFile            = "peers.json"
MainNetworkPort      = 8108
MainSeedURL          = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt"
MainSpecialPeers     = ""
TestNetworkPort      = 8109
TestSeedURL          = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/testseed.txt"
TestSpecialPeers     = ""
LocalNetworkPort     = 8110
LocalSeedURL         = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/localseed.txt"
LocalSpecialPeers    = ""
CustomNetworkPort    = 8110
CustomSeedURL        = ""
CustomSpecialPeers   = ""
CustomBootstrapIdentity     = 38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9
CustomBootstrapKey          = cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a
; The maximum number of other peers dialing into this node that will be accepted
P2PIncoming	= 200
; The maximum number of peers this node will attempt to dial into
P2POutgoing	= 32
; --------------- NodeMode: FULL | SERVER ----------------
NodeMode                                = FULL
LocalServerPrivKey                      = 4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d
LocalServerPublicKey                    = cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a
ExchangeRateChainId                     = 111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03
ExchangeRateAuthorityPublicKeyMainNet   = daf5815c2de603dbfa3e1e64f88a5cf06083307cf40da4a9b539c41832135b4a
ExchangeRateAuthorityPublicKeyTestNet   = 1d75de249c2fc0384fb6701b30dc86b39dc72e5a47ba4f79ef250d39e21e7a4f
; Private key all zeroes:
ExchangeRateAuthorityPublicKeyLocalNet  = 3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29

; These define if the RPC and Control Panel connection to factomd should be encrypted, and if it is, what files
; are the secret key and the public certificate.  factom-cli and factom-walletd uses the certificate specified here if TLS is enabled.
; To use default files and paths leave /full/path/to/... in place.
FactomdTlsEnabled                     = false
FactomdTlsPrivateKey                  = "/full/path/to/factomdAPIpriv.key"
FactomdTlsPublicCert                  = "/full/path/to/factomdAPIpub.cert"

; These are the username and password that factomd requires for the RPC API and the Control Panel
; This file is also used by factom-cli and factom-walletd to determine what login to use
FactomdRpcUser                        = ""
FactomdRpcPass                        = ""

; RequestTimeout is the amount of time in seconds before a pending request for a
; missing DBState is considered too old and the state is put back into the
; missing states list. If RequestTimout is not set or is set to 0 it will become
; 1/10th of DirectoryBlockInSeconds
;RequestTimeout						= 30
; RequestLimit is the maximum number of pending requests for missing states.
; factomd will stop making DBStateMissing requests until current requests are
; moved out of the waiting list
RequestLimit						= 200

; This paramater allows Cross-Origin Resource Sharing (CORS) so web browsers will use data returned from the API when called from the listed URLs
; Example paramaters are "http://www.example.com, http://anotherexample.com, *"
CorsDomains                           = ""

; Specifying when to change ACKs for switching leader servers
ChangeAcksHeight                      = 0

; ------------------------------------------------------------------------------
; logLevel - allowed values are: debug, info, notice, warning, error, critical, alert, emergency and none
; ConsoleLogLevel - allowed values are: debug, standard
; ------------------------------------------------------------------------------
[log]
logLevel                              = error
LogPath                               = "database/Log"
ConsoleLogLevel                       = standard

; ------------------------------------------------------------------------------
; Configurations for factom-walletd
; ------------------------------------------------------------------------------
[Walletd]
; These are the username and password that factom-walletd requires
; This file is also used by factom-cli to determine what login to use
WalletRpcUser                         = ""
WalletRpcPass                         = ""

; These define if the connection to the wallet should be encrypted, and if it is, what files
; are the secret key and the public certificate.  factom-cli uses the certificate specified here if TLS is enabled.
; To use default files and paths leave /full/path/to/... in place.
WalletTlsEnabled                      = false
WalletTlsPrivateKey                   = "/full/path/to/walletAPIpriv.key"
WalletTlsPublicCert                   = "/full/path/to/walletAPIpub.cert"

; This is where factom-walletd and factom-cli will find factomd to interact with the blockchain
; This value can also be updated to authorize an external ip or domain name when factomd creates a TLS cert
FactomdLocation                       = "localhost:8088"

; This is where factom-cli will find factom-walletd to create Factoid and Entry Credit transactions
; This value can also be updated to authorize an external ip or domain name when factom-walletd creates a TLS cert
WalletdLocation                       = "localhost:8089"

; Enables wallet database encryption on factom-walletd. If this option is enabled, an unencrypted database
; cannot exist. If an unencrypted database exists, the wallet will exit.
WalletEncrypted                       = false

; ------------------------------------------------------------------------------
; Configuration options for the live feed API
; ------------------------------------------------------------------------------
[LiveFeedAPI]
EnableLiveFeedAPI                     = false
EventReceiverProtocol                 = tcp
EventReceiverAddress                  = 127.0.0.1
EventReceiverPort                     = 8040
EventFormat                           = protobuf
`

func (s *FactomdConfig) String() string {
	var out primitives.Buffer

	out.WriteString(fmt.Sprintf("\nFactomd Config"))
	out.WriteString(fmt.Sprintf("\n  App"))
	out.WriteString(fmt.Sprintf("\n    PortNumber              %v", s.App.PortNumber))
	out.WriteString(fmt.Sprintf("\n    HomeDir                 %v", s.App.HomeDir))
	out.WriteString(fmt.Sprintf("\n    ControlPanelPort        %v", s.App.ControlPanelPort))
	out.WriteString(fmt.Sprintf("\n    ControlPanelFilesPath   %v", s.App.ControlPanelFilesPath))
	out.WriteString(fmt.Sprintf("\n    ControlPanelSetting     %v", s.App.ControlPanelSetting))
	out.WriteString(fmt.Sprintf("\n    DBType                  %v", s.App.DBType))
	out.WriteString(fmt.Sprintf("\n    LdbPath                 %v", s.App.LdbPath))
	out.WriteString(fmt.Sprintf("\n    BoltDBPath              %v", s.App.BoltDBPath))
	out.WriteString(fmt.Sprintf("\n    DataStorePath           %v", s.App.DataStorePath))
	out.WriteString(fmt.Sprintf("\n    DirectoryBlockInSeconds %v", s.App.DirectoryBlockInSeconds))
	out.WriteString(fmt.Sprintf("\n    ExportData              %v", s.App.ExportData))
	out.WriteString(fmt.Sprintf("\n    ExportDataSubpath       %v", s.App.ExportDataSubpath))
	out.WriteString(fmt.Sprintf("\n    Network                 %v", s.App.Network))
	out.WriteString(fmt.Sprintf("\n    MainNetworkPort         %v", s.App.MainNetworkPort))
	out.WriteString(fmt.Sprintf("\n    PeersFile               %v", s.App.PeersFile))
	out.WriteString(fmt.Sprintf("\n    MainSeedURL             %v", s.App.MainSeedURL))
	out.WriteString(fmt.Sprintf("\n    MainSpecialPeers        %v", s.App.MainSpecialPeers))
	out.WriteString(fmt.Sprintf("\n    TestNetworkPort         %v", s.App.TestNetworkPort))
	out.WriteString(fmt.Sprintf("\n    TestSeedURL             %v", s.App.TestSeedURL))
	out.WriteString(fmt.Sprintf("\n    TestSpecialPeers        %v", s.App.TestSpecialPeers))
	out.WriteString(fmt.Sprintf("\n    LocalNetworkPort        %v", s.App.LocalNetworkPort))
	out.WriteString(fmt.Sprintf("\n    LocalSeedURL            %v", s.App.LocalSeedURL))
	out.WriteString(fmt.Sprintf("\n    LocalSpecialPeers       %v", s.App.LocalSpecialPeers))
	out.WriteString(fmt.Sprintf("\n    CustomNetworkPort       %v", s.App.CustomNetworkPort))
	out.WriteString(fmt.Sprintf("\n    CustomSeedURL           %v", s.App.CustomSeedURL))
	out.WriteString(fmt.Sprintf("\n    CustomSpecialPeers      %v", s.App.CustomSpecialPeers))
	out.WriteString(fmt.Sprintf("\n    CustomBootstrapIdentity %v", s.App.CustomBootstrapIdentity))
	out.WriteString(fmt.Sprintf("\n    CustomBootstrapKey      %v", s.App.CustomBootstrapKey))
	out.WriteString(fmt.Sprintf("\n    P2PIncoming             %v", s.App.P2PIncoming))
	out.WriteString(fmt.Sprintf("\n    P2POutgoing             %v", s.App.P2POutgoing))
	out.WriteString(fmt.Sprintf("\n    NodeMode                %v", s.App.NodeMode))
	out.WriteString(fmt.Sprintf("\n    IdentityChainID         %v", s.App.IdentityChainID))
	out.WriteString(fmt.Sprintf("\n    LocalServerPrivKey      %v", s.App.LocalServerPrivKey))
	out.WriteString(fmt.Sprintf("\n    LocalServerPublicKey    %v", s.App.LocalServerPublicKey))
	out.WriteString(fmt.Sprintf("\n    ExchangeRate            %v", s.App.ExchangeRate))
	out.WriteString(fmt.Sprintf("\n    ExchangeRateChainId     %v", s.App.ExchangeRateChainId))
	out.WriteString(fmt.Sprintf("\n    ExchangeRateAuthorityPublicKey   %v", s.App.ExchangeRateAuthorityPublicKey))
	out.WriteString(fmt.Sprintf("\n    FactomdTlsEnabled        %v", s.App.FactomdTlsEnabled))
	out.WriteString(fmt.Sprintf("\n    FactomdTlsPrivateKey     %v", s.App.FactomdTlsPrivateKey))
	out.WriteString(fmt.Sprintf("\n    FactomdTlsPublicCert     %v", s.App.FactomdTlsPublicCert))
	out.WriteString(fmt.Sprintf("\n    FactomdRpcUser          	%v", s.App.FactomdRpcUser))
	out.WriteString(fmt.Sprintf("\n    FactomdRpcPass          	%v", s.App.FactomdRpcPass))
	out.WriteString(fmt.Sprintf("\n    ChangeAcksHeight         %v", s.App.ChangeAcksHeight))
	out.WriteString(fmt.Sprintf("\n    BitcoinAnchorRecordPublicKeys    %v", s.App.BitcoinAnchorRecordPublicKeys))
	out.WriteString(fmt.Sprintf("\n    EthereumAnchorRecordPublicKeys    %v", s.App.EthereumAnchorRecordPublicKeys))

	out.WriteString(fmt.Sprintf("\n  Log"))
	out.WriteString(fmt.Sprintf("\n    LogPath                 %v", s.Log.LogPath))
	out.WriteString(fmt.Sprintf("\n    LogLevel                %v", s.Log.LogLevel))
	out.WriteString(fmt.Sprintf("\n    ConsoleLogLevel         %v", s.Log.ConsoleLogLevel))

	out.WriteString(fmt.Sprintf("\n  Walletd"))
	out.WriteString(fmt.Sprintf("\n    WalletRpcUser           %v", s.Walletd.WalletRpcUser))
	out.WriteString(fmt.Sprintf("\n    WalletRpcPass           %v", s.Walletd.WalletRpcPass))
	out.WriteString(fmt.Sprintf("\n    WalletTlsEnabled        %v", s.Walletd.WalletTlsEnabled))
	out.WriteString(fmt.Sprintf("\n    WalletTlsPrivateKey     %v", s.Walletd.WalletTlsPrivateKey))
	out.WriteString(fmt.Sprintf("\n    WalletTlsPublicCert     %v", s.Walletd.WalletTlsPublicCert))
	out.WriteString(fmt.Sprintf("\n    FactomdLocation         %v", s.Walletd.FactomdLocation))
	out.WriteString(fmt.Sprintf("\n    WalletdLocation         %v", s.Walletd.WalletdLocation))
	out.WriteString(fmt.Sprintf("\n    WalletEncryption        %v", s.Walletd.WalletEncrypted))

	out.WriteString(fmt.Sprintf("\n  LiveFeedAPI"))
	out.WriteString(fmt.Sprintf("\n    EnableLiveFeedAPI       %v", s.LiveFeedAPI.EnableLiveFeedAPI))
	out.WriteString(fmt.Sprintf("\n    EventReceiverProtocol   %v", s.LiveFeedAPI.EventReceiverProtocol))
	out.WriteString(fmt.Sprintf("\n    EventReceiverAddress    %v", s.LiveFeedAPI.EventReceiverAddress))
	out.WriteString(fmt.Sprintf("\n    EventReceiverPort       %v", s.LiveFeedAPI.EventReceiverPort))
	out.WriteString(fmt.Sprintf("\n    EventFormat             %v", s.LiveFeedAPI.EventFormat))

	return out.String()
}

func ConfigFilename() string {
	return GetHomeDir() + "/.factom/m2/factomd.conf"
}

func GetConfigFilename(dir string) string {
	return GetHomeDir() + "/.factom/" + dir + "/factomd.conf"
}

func GetChangeAcksHeight(filename string) (change uint32, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error getting acks - %v\n", r)
		}
	}()

	config := ReadConfig(filename)

	return config.App.ChangeAcksHeight, nil
}

// Check for absolute path on Windows or linux or Darwin(Mac)
var pathPresent *regexp.Regexp

func CheckConfigFileName(filename string) string {
	// compile the regex if this is the first time.
	if pathPresent == nil {
		var err error
		// paths may look like C: or c: or ~/ or ./ or ../ or / or \ at the start of a filename
		pathPresent, err = regexp.Compile(`^([A-Za-z]:)|(~?(\.\.?)?[/\\])`)
		if err != nil {
			panic(err)
		}
	}
	// Check for absolute path on Windows or linux or Darwin(Mac)
	// if path is relative prepend the Factom Home path
	if !pathPresent.MatchString(filename) {
		filename = GetHomeDir() + "/.factom/m2/" + filename
	}
	return filename
}

// Track a filename-error pair so we don't report the same error repeatedly
var reportedError map[string]string = make(map[string]string)

func ReadConfig(filename string) *FactomdConfig {
	if filename == "" {
		filename = ConfigFilename()
	}
	filename = CheckConfigFileName(filename)

	cfg := new(FactomdConfig)

	err := gcfg.ReadStringInto(cfg, defaultConfig)
	if err != nil {
		panic(err)
	}

	err = gcfg.FatalOnly(gcfg.ReadFileInto(cfg, filename))
	if err != nil {
		if reportedError[filename] != err.Error() {
			log.Printfln("Reading from '%s'", filename)
			log.Printfln("Cannot open custom config file,\nStarting with default settings.\n%v\n", err)
			// Remember the error reported for this filename
			reportedError[filename] = err.Error()
		}

		err = gcfg.ReadStringInto(cfg, defaultConfig)
		if err != nil {
			panic(err)
		}
	} else {
		// Remember that there was no error reported for this filename
		delete(reportedError, filename)
	}

	// Default to home directory if not set
	if len(cfg.App.HomeDir) < 1 {
		cfg.App.HomeDir = GetHomeDir() + "/.factom/m2/"
	} else {
		cfg.App.HomeDir = cfg.App.HomeDir + "/.factom/m2/"
	}

	if len(cfg.App.FastBootLocation) < 1 {
		cfg.App.FastBootLocation = cfg.App.HomeDir
	}

	switch cfg.App.Network {
	case "MAIN":
		cfg.App.ExchangeRateAuthorityPublicKey = cfg.App.ExchangeRateAuthorityPublicKeyMainNet
		break
	case "TEST":
		cfg.App.ExchangeRateAuthorityPublicKey = cfg.App.ExchangeRateAuthorityPublicKeyTestNet
		break
	case "LOCAL":
		cfg.App.ExchangeRateAuthorityPublicKey = cfg.App.ExchangeRateAuthorityPublicKeyLocalNet
		break
	}

	if len(cfg.App.BitcoinAnchorRecordPublicKeys) == 0 {
		cfg.App.BitcoinAnchorRecordPublicKeys = []string{
			"0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a", // m1 key
			"d569419348ed7056ec2ba54f0ecd9eea02648b260b26e0474f8c07fe9ac6bf83", // m2 key
		}
	}
	if len(cfg.App.EthereumAnchorRecordPublicKeys) == 0 {
		cfg.App.EthereumAnchorRecordPublicKeys = []string{
			"a4a7905ab2226f267c6b44e1d5db2c97638b7bbba72fd1823d053ccff2892455",
		}
	}

	return cfg
}

func GetHomeDir() string {
	factomhome := os.Getenv("FACTOM_HOME")
	if factomhome != "" {
		return factomhome
	}

	// Get the OS specific home directory via the Go standard lib.
	var homeDir string
	usr, err := user.Current()
	if err == nil {
		homeDir = usr.HomeDir
	}

	// Fall back to standard HOME environment variable that works
	// for most POSIX OSes if the directory from the Go standard
	// lib failed.
	if err != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}
