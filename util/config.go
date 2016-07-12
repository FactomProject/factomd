package util

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"

	"gopkg.in/gcfg.v1"
)

var _ = fmt.Print

type FactomdConfig struct {
	App struct {
		PortNumber                   int
		HomeDir                      string
		DBType                       string
		LdbPath                      string
		BoltDBPath                   string
		DataStorePath                string
		DirectoryBlockInSeconds      int
		ExportData                   bool
		ExportDataSubpath            string
		Network                      string
		PeersFile                    string
		SeedURL                      string
		NodeMode                     string
		LocalServerPrivKey           string
		LocalServerPublicKey         string
		ExchangeRate                 uint64
		ExchangeRateChainId          string
		ExchangeRateAuthorityAddress string
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
	Anchor struct {
		ServerECPrivKey     string
		ServerECPublicKey   string
		AnchorChainID       string
		ConfirmationsNeeded int
	}
	Btc struct {
		BTCPubAddr         string
		SendToBTCinSeconds int
		WalletPassphrase   string
		CertHomePath       string
		RpcClientHost      string
		RpcClientEndpoint  string
		RpcClientUser      string
		RpcClientPass      string
		BtcTransFee        float64
		CertHomePathBtcd   string
		RpcBtcdHost        string
	}
	Rpc struct {
		PortNumber       int
		ApplicationName  string
		RefreshInSeconds int
	}
	Wsapi struct {
		PortNumber      int
		ApplicationName string
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
}

// defaultConfig
const defaultConfig = `
; ------------------------------------------------------------------------------
; App settings
; ------------------------------------------------------------------------------
[app]
PortNumber                            = 8088
HomeDir                               = ""
; --------------- DBType: LDB | Bolt | Map
DBType                                = "Map"
LdbPath                               = "database/ldb"
BoltDBPath                            = "database/bolt"
DataStorePath                         = "data/export"
DirectoryBlockInSeconds               = 6
ExportData                            = true
ExportDataSubpath                     = "database/export/"
; --------------- Network: MAIN | TEST | LOCAL
Network                               = LOCAL
PeersFile                             = "peers.json"
SeedURL                               = "http://factomstatus.com/seed/seed.txt"
; --------------- NodeMode: FULL | SERVER | LIGHT ----------------
NodeMode                              = FULL
LocalServerPrivKey                    = 4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d
LocalServerPublicKey                  = cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a
ExchangeRate                          = 00100000
ExchangeRateChainId		      = eac57815972c504ec5ae3f9e5c1fe12321a3c8c78def62528fb74cf7af5e7389
ExchangeRateAuthorityAddress          = EC2DKSYyRcNWf7RS963VFYgMExoHRYLHVeCfQ9PGPmNzwrcmgm2r

[anchor]
ServerECPrivKey                       = 397c49e182caa97737c6b394591c614156fbe7998d7bf5d76273961e9fa1edd4
ServerECPublicKey                     = 06ed9e69bfdf85db8aa69820f348d096985bc0b11cc9fc9dcee3b8c68b41dfd5
AnchorChainID                         = df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604
ConfirmationsNeeded                   = 20

[btc]
WalletPassphrase                      = "lindasilva"
CertHomePath                          = "btcwallet"
RpcClientHost                         = "localhost:18332"
RpcClientEndpoint                     = "ws"
RpcClientUser                         = "testuser"
RpcClientPass                         = "notarychain"
BtcTransFee                           = 0.000001
CertHomePathBtcd                      = "btcd"
RpcBtcdHost                           = "localhost:18334"

[wsapi]
ApplicationName                       = "Factom/wsapi"
PortNumber                            = 8088

; ------------------------------------------------------------------------------
; logLevel - allowed values are: debug, info, notice, warning, error, critical, alert, emergency and none
; ConsoleLogLevel - allowed values are: debug, standard
; ------------------------------------------------------------------------------
[log]
logLevel                              = error
LogPath                               = "database/"
ConsoleLogLevel                       = standard

; ------------------------------------------------------------------------------
; Configurations for fctwallet
; ------------------------------------------------------------------------------
[Wallet]
Address                               = localhost
Port                                  = 8089
DataFile                              = fctwallet.dat
RefreshInSeconds                      = 6
BoltDBPath                            = ""
`

func (s *FactomdConfig) String() string {
	var out primitives.Buffer

	out.WriteString(fmt.Sprintf("\nFactomd Config"))
	out.WriteString(fmt.Sprintf("\n  App"))
	out.WriteString(fmt.Sprintf("\n    PortNumber              %v", s.App.PortNumber))
	out.WriteString(fmt.Sprintf("\n    HomeDir                 %v", s.App.HomeDir))
	out.WriteString(fmt.Sprintf("\n    DBType                 %v", s.App.DBType))
	out.WriteString(fmt.Sprintf("\n    LdbPath                 %v", s.App.LdbPath))
	out.WriteString(fmt.Sprintf("\n    BoltDBPath              %v", s.App.BoltDBPath))
	out.WriteString(fmt.Sprintf("\n    DataStorePath           %v", s.App.DataStorePath))
	out.WriteString(fmt.Sprintf("\n    DirectoryBlockInSeconds %v", s.App.DirectoryBlockInSeconds))
	out.WriteString(fmt.Sprintf("\n    ExportData              %v", s.App.ExportData))
	out.WriteString(fmt.Sprintf("\n    ExportDataSubpath       %v", s.App.ExportDataSubpath))
	out.WriteString(fmt.Sprintf("\n    Network                 %v", s.App.Network))
	out.WriteString(fmt.Sprintf("\n    PeersFile               %v", s.App.PeersFile))
	out.WriteString(fmt.Sprintf("\n    SeedURL                 %v", s.App.SeedURL))
	out.WriteString(fmt.Sprintf("\n    NodeMode                %v", s.App.NodeMode))
	out.WriteString(fmt.Sprintf("\n    LocalServerPrivKey      %v", s.App.LocalServerPrivKey))
	out.WriteString(fmt.Sprintf("\n    LocalServerPublicKey    %v", s.App.LocalServerPublicKey))
	out.WriteString(fmt.Sprintf("\n    ExchangeRate            %v", s.App.ExchangeRate))
	out.WriteString(fmt.Sprintf("\n    ExchangeRateChainId     %v", s.App.ExchangeRateChainId))
	out.WriteString(fmt.Sprintf("\n    ExchangeRateAuthorityAddress   %v", s.App.ExchangeRateAuthorityAddress))

	out.WriteString(fmt.Sprintf("\n  Anchor"))
	out.WriteString(fmt.Sprintf("\n    ServerECPrivKey         %v", s.Anchor.ServerECPrivKey))
	out.WriteString(fmt.Sprintf("\n    ServerECPublicKey       %v", s.Anchor.ServerECPublicKey))
	out.WriteString(fmt.Sprintf("\n    AnchorChainID           %v", s.Anchor.AnchorChainID))
	out.WriteString(fmt.Sprintf("\n    ConfirmationsNeeded     %v", s.Anchor.ConfirmationsNeeded))

	out.WriteString(fmt.Sprintf("\n  Btc"))
	out.WriteString(fmt.Sprintf("\n    BTCPubAddr              %v", s.Btc.BTCPubAddr))
	out.WriteString(fmt.Sprintf("\n    SendToBTCinSeconds      %v", s.Btc.SendToBTCinSeconds))
	out.WriteString(fmt.Sprintf("\n    WalletPassphrase        %v", s.Btc.WalletPassphrase))
	out.WriteString(fmt.Sprintf("\n    CertHomePath            %v", s.Btc.CertHomePath))
	out.WriteString(fmt.Sprintf("\n    RpcClientHost           %v", s.Btc.RpcClientHost))
	out.WriteString(fmt.Sprintf("\n    RpcClientEndpoint       %v", s.Btc.RpcClientEndpoint))
	out.WriteString(fmt.Sprintf("\n    RpcClientUser           %v", s.Btc.RpcClientUser))
	out.WriteString(fmt.Sprintf("\n    RpcClientPass           %v", s.Btc.RpcClientPass))
	out.WriteString(fmt.Sprintf("\n    BtcTransFee             %v", s.Btc.BtcTransFee))
	out.WriteString(fmt.Sprintf("\n    CertHomePathBtcd        %v", s.Btc.CertHomePathBtcd))
	out.WriteString(fmt.Sprintf("\n    RpcBtcdHost             %v", s.Btc.RpcBtcdHost))

	out.WriteString(fmt.Sprintf("\n  Rpc"))
	out.WriteString(fmt.Sprintf("\n    PortNumber              %v", s.Rpc.PortNumber))
	out.WriteString(fmt.Sprintf("\n    ApplicationName         %v", s.Rpc.ApplicationName))
	out.WriteString(fmt.Sprintf("\n    RefreshInSeconds        %v", s.Rpc.RefreshInSeconds))

	out.WriteString(fmt.Sprintf("\n  Wsapi"))
	out.WriteString(fmt.Sprintf("\n    PortNumber              %v", s.Wsapi.PortNumber))
	out.WriteString(fmt.Sprintf("\n    ApplicationName         %v", s.Wsapi.ApplicationName))

	out.WriteString(fmt.Sprintf("\n  Log"))
	out.WriteString(fmt.Sprintf("\n    LogPath                 %v", s.Log.LogPath))
	out.WriteString(fmt.Sprintf("\n    LogLevel                %v", s.Log.LogLevel))
	out.WriteString(fmt.Sprintf("\n    ConsoleLogLevel         %v", s.Log.ConsoleLogLevel))

	out.WriteString(fmt.Sprintf("\n  Wallet"))
	out.WriteString(fmt.Sprintf("\n    Address                 %v", s.Wallet.Address))
	out.WriteString(fmt.Sprintf("\n    Port                    %v", s.Wallet.Port))
	out.WriteString(fmt.Sprintf("\n    DataFile                %v", s.Wallet.DataFile))
	out.WriteString(fmt.Sprintf("\n    RefreshInSeconds        %v", s.Wallet.RefreshInSeconds))
	out.WriteString(fmt.Sprintf("\n    BoltDBPath              %v", s.Wallet.BoltDBPath))

	return out.String()
}

func ConfigFilename() string {
	return GetHomeDir() + "/.factom/m2/factomd.conf"
}

func GetConfigFilename(dir string) string {
	return GetHomeDir() + "/.factom/" + dir + "/factomd.conf"
}

func ReadConfig(filename string, folder string) *FactomdConfig {
	if filename == "" {
		filename = ConfigFilename()
	}
	cfg := new(FactomdConfig)

	err := gcfg.ReadStringInto(cfg, defaultConfig)
	if err != nil {
		panic(err)
	}

	err = gcfg.ReadFileInto(cfg, filename)
	if err != nil {
		log.Printfln("Reading from '%s'", filename)
		log.Printfln("ERROR Reading config file!\nServer starting with default settings...\n%v\n", err)
		err = gcfg.ReadStringInto(cfg, defaultConfig)
		if err != nil {
			panic(err)
		}
	}

	// Default to home directory if not set
	if len(cfg.App.HomeDir) < 1 {
		cfg.App.HomeDir = GetHomeDir() + "/.factom/m2/"
	} else {
		cfg.App.HomeDir = cfg.App.HomeDir + "/.factom/m2/"
	}

	// TODO: improve the paths after milestone 1
	cfg.App.LdbPath = cfg.App.HomeDir + folder + cfg.App.LdbPath
	cfg.App.BoltDBPath = cfg.App.HomeDir + folder + cfg.App.BoltDBPath
	cfg.App.DataStorePath = cfg.App.HomeDir + folder + cfg.App.DataStorePath
	cfg.Log.LogPath = cfg.App.HomeDir + folder + cfg.Log.LogPath
	cfg.Wallet.BoltDBPath = cfg.App.HomeDir + folder + cfg.Wallet.BoltDBPath
	cfg.App.ExportDataSubpath = cfg.App.HomeDir + folder + cfg.App.ExportDataSubpath
	cfg.App.PeersFile = cfg.App.HomeDir + cfg.App.PeersFile

	return cfg
}

func GetHomeDir() string {
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
