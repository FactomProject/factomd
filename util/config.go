package util

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/log"
	"os"
	"os/user"

	"code.google.com/p/gcfg"
)

var _ = fmt.Print

type FactomdConfig struct {
	App struct {
		PortNumber              int
		HomeDir                 string
		DBType                  string
		LdbPath                 string
		BoltDBPath              string
		DataStorePath           string
		DirectoryBlockInSeconds int
		Network                 string
		NodeMode                string
		LocalServerPrivKey      string
		LocalServerPublicKey    string
		ExchangeRate            uint64
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
		RpcUser            string
		RpcPass            string
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
	}

	//    AddPeers     []string `short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
	//    ConnectPeers []string `long:"connect" description:"Connect only to the specified peers at startup"`

	Proxy          string `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	DisableListen  bool   `long:"nolisten" description:"Disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen"`
	DisableRPC     bool   `long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass is specified"`
	DisableTLS     bool   `long:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
	DisableDNSSeed bool   `long:"nodnsseed" description:"Disable DNS seeding for peers"`
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
DBType                                = "Bolt"
LdbPath                               = "ldb"
BoltDBPath                            = "bolt"
DataStorePath                         = "data/export/"
DirectoryBlockInSeconds               = 600
; --------------- Network: MAIN | TEST | LOCAL 
Network                               = LOCAL
; --------------- NodeMode: FULL | SERVER | LIGHT ----------------
NodeMode                              = FULL
LocalServerPrivKey                    = 4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d
LocalServerPublicKey                  = cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a
ExchangeRate                          = 00100000

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
BtcTransFee                           = 0.0001
CertHomePathBtcd                      = "btcd"
RpcBtcdHost                           = "localhost:18334"
RpcUser                               =testuser
RpcPass                               =notarychain

[wsapi]
ApplicationName                       = "Factom/wsapi"
PortNumber                            =8088

; ------------------------------------------------------------------------------
; logLevel - allowed values are: debug, info, notice, warning, error, critical, alert, emergency and none
; ConsoleLogLevel - allowed values are: debug, standard
; ------------------------------------------------------------------------------
[log]
logLevel                              =debug
LogPath                               = "m2factom-d.log"
ConsoleLogLevel                       =debug

; ------------------------------------------------------------------------------
; Configurations for fctwallet
; ------------------------------------------------------------------------------
[Wallet]
Address                               = localhost
Port                                  = 8089
DataFile                              = fctwallet.dat
RefreshInSeconds                      = 60
BoltDBPath                            = ""
`

func (s *FactomdConfig) String() string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("\nFactomd Config"))
	out.WriteString(fmt.Sprintf("\n  App"))
	out.WriteString(fmt.Sprintf("\n    PortNumber              %v", s.App.PortNumber))
	out.WriteString(fmt.Sprintf("\n    HomeDir                 %v", s.App.HomeDir))
	out.WriteString(fmt.Sprintf("\n    DBType                 %v", s.App.DBType))
	out.WriteString(fmt.Sprintf("\n    LdbPath                 %v", s.App.LdbPath))
	out.WriteString(fmt.Sprintf("\n    BoltDBPath              %v", s.App.BoltDBPath))
	out.WriteString(fmt.Sprintf("\n    DataStorePath           %v", s.App.DataStorePath))
	out.WriteString(fmt.Sprintf("\n    DirectoryBlockInSeconds %v", s.App.DirectoryBlockInSeconds))
	out.WriteString(fmt.Sprintf("\n    Network                 %v", s.App.Network))
	out.WriteString(fmt.Sprintf("\n    NodeMode                %v", s.App.NodeMode))
	out.WriteString(fmt.Sprintf("\n    LocalServerPrivKey      %v", s.App.LocalServerPrivKey))
	out.WriteString(fmt.Sprintf("\n    LocalServerPublicKey    %v", s.App.LocalServerPublicKey))
	out.WriteString(fmt.Sprintf("\n    ExchangeRate            %v", s.App.ExchangeRate))

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
	out.WriteString(fmt.Sprintf("\n    RpcUser                 %v", s.Btc.RpcUser))
	out.WriteString(fmt.Sprintf("\n    RpcPass                 %v", s.Btc.RpcPass))

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

func ReadConfig(filename string) *FactomdConfig {
	if filename == "" {
		filename = ConfigFilename()
	}
	cfg := new(FactomdConfig)

	// This makes factom config file located at
	//   POSIX (Linux/BSD): ~/.factom/factom.conf
	//   Mac OS: $HOME/Library/Application Support/Factom/factom.conf
	//   Windows: %LOCALAPPDATA%\Factom\factom.conf
	//   Plan 9: $home/factom/factom.conf
	//factomHomeDir := btcutil.AppDataDir("factom", false)
	//defaultConfigFile := filepath.Join(factomHomeDir, "factomd.conf")
	//
	// eventually we need to make data dir as following
	//defaultDataDir   = filepath.Join(factomHomeDir, "data")
	//LdbPath                     = filepath.Join(defaultDataDir, "ldb9")
	//DataStorePath         = filepath.Join(defaultDataDir, "store/seed/")

	err := gcfg.ReadFileInto(cfg, filename)
	if err != nil {
		log.Printfln("Reading from '%s'", filename)
		log.Printfln("ERROR Reading config file!\nServer starting with default settings...\n%v\n", err)
		gcfg.ReadStringInto(cfg, defaultConfig)
	}

	// Default to home directory if not set
	if len(cfg.App.HomeDir) < 1 {
		cfg.App.HomeDir = GetHomeDir() + "/.factom/m2/"
	}

	// TODO: improve the paths after milestone 1
	cfg.App.LdbPath = cfg.App.HomeDir + cfg.App.LdbPath
	cfg.App.BoltDBPath = cfg.App.HomeDir + cfg.App.BoltDBPath
	cfg.App.DataStorePath = cfg.App.HomeDir + cfg.App.DataStorePath
	cfg.Log.LogPath = cfg.App.HomeDir + cfg.Log.LogPath
	cfg.Wallet.BoltDBPath = cfg.App.HomeDir + cfg.Wallet.BoltDBPath

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
