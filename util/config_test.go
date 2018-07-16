package util_test

import (
	"testing"

	. "github.com/FactomProject/factomd/util"
	"gopkg.in/gcfg.v1"
)

func TestLoadDefaultConfig(t *testing.T) {
	type testConfig struct {
		Test struct {
			Foo string
			Bar int64
		}
	}

	var testConfigFile string = `
	[Test]
	Foo = "Bla"
	Bar = "-1"
	`

	cfg := new(testConfig)
	gcfg.ReadStringInto(cfg, testConfigFile)
	if cfg.Test.Foo != "Bla" {
		t.Errorf("Wrong variable read - %v", cfg.Test.Foo)
	}
	if cfg.Test.Bar != -1 {
		t.Errorf("Wrong variable read - %v", cfg.Test.Bar)
	}

	var testConfigFile2 string = `
	[Test]
	Foo = "Ble"
	`
	cfg2 := new(testConfig)
	gcfg.ReadStringInto(cfg2, testConfigFile2)
	if cfg2.Test.Foo != "Ble" {
		t.Errorf("Wrong variable read - %v", cfg.Test.Foo)
	}
	if cfg2.Test.Bar != 0 {
		t.Errorf("Wrong variable read - %v", cfg.Test.Bar)
	}

	gcfg.ReadStringInto(cfg, testConfigFile2)
	if cfg.Test.Foo != "Ble" {
		t.Errorf("Wrong variable read - %v", cfg.Test.Foo)
	}
	if cfg.Test.Bar != -1 {
		t.Errorf("Wrong variable read - %v", cfg.Test.Bar)
	}
}

func TestLoadDefaultConfigFull(t *testing.T) {
	var defaultConfig string = `
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
DBType                                = "Map"
LdbPath                               = "ldb"
BoltDBPath                            = ""
DataStorePath                         = "data/export/"
DirectoryBlockInSeconds               = 6
ExportData                            = false
ExportDataSubpath                     = "database/export/"
; --------------- Network: MAIN | TEST | LOCAL
Network                               = LOCAL
PeersFile               = "peers.json"
MainNetworkPort         = 8108
MainSeedURL             = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt"
MainSpecialPeers        = ""
TestNetworkPort         = 8109
TestSeedURL             = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/testseed.txt"
TestSpecialPeers        = ""
LocalNetworkPort        = 8110
LocalSeedURL            = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/localseed.txt"
LocalSpecialPeers       = ""
CustomNetworkPort       = 8110
CustomSeedURL           = ""
CustomSpecialPeers      = ""
CustomBootstrapIdentity = 38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9
CustomBootstrapKey      = cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a
; --------------- NodeMode: FULL | SERVER | LIGHT ----------------
NodeMode                              = FULL
LocalServerPrivKey                    = 4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d
LocalServerPublicKey                  = cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a
ExchangeRateChainId                   = 111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03
ExchangeRateAuthorityAddress          = EC2DKSYyRcNWf7RS963VFYgMExoHRYLHVeCfQ9PGPmNzwrcmgm2r

; These define if the PRC and Control Panel connection to factomd should be encrypted, and if it is, what files
; are the secret key and the public certificate.  factom-cli and factom-walletd uses the certificate specified here if TLS is enabled.
FactomdTlsEnabled                     = false
FactomdTlsPrivateKey                  = "/full/path/to/factomdAPIpriv.key"
FactomdTlsPublicCert                  = "/full/path/to/factomdAPIpub.cert"

; These are the username and password that factomd requires for the RPC API and the Control Panel
; This file is also used by factom-cli and factom-walletd to determine what login to use
FactomdRpcUser                        = ""
FactomdRpcPass                        = ""

; Specifying when to stop or start ACKs for switching leader servers
ChangeAcksHeight                      = 123

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
WalletTlsEnabled                      = false
WalletTlsPrivateKey                   = "/full/path/to/walletAPIpriv.key"
WalletTlsPublicCert                   = "/full/path/to/walletAPIpub.cert"

; This is where factom-walletd and factom-cli will find factomd to interact with the blockchain
; This value can also be updated to authorize an external ip or domain name when factomd creates a TLS cert
FactomdLocation                       = "localhost:8088"

; This is where factom-cli will find factom-walletd to create Factoid and Entry Credit transactions
; This value can also be updated to authorize an external ip or domain name when factom-walletd creates a TLS cert
WalletdLocation                       = "localhost:8089"
	`

	var modifiedConfig string = `
	[app]
	DBType                                = "MapMap"
	LdbPath                               = ""
	BoltDBPath                            = "Something"
	`

	cfg := new(FactomdConfig)
	gcfg.ReadStringInto(cfg, defaultConfig)
	if cfg.App.DBType != "Map" {
		t.Errorf("Wrong variable read - %v", cfg.App.DBType)
	}
	if cfg.App.LdbPath != "ldb" {
		t.Errorf("Wrong variable read - %v", cfg.App.LdbPath)
	}
	if cfg.App.BoltDBPath != "" {
		t.Errorf("Wrong variable read - %v", cfg.App.BoltDBPath)
	}
	if cfg.App.DataStorePath != "data/export/" {
		t.Errorf("Wrong variable read - %v", cfg.App.DataStorePath)
	}

	gcfg.ReadStringInto(cfg, modifiedConfig)
	if cfg.App.DBType != "MapMap" {
		t.Errorf("Wrong variable read - %v", cfg.App.DBType)
	}
	if cfg.App.LdbPath != "" {
		t.Errorf("Wrong variable read - %v", cfg.App.LdbPath)
	}
	if cfg.App.BoltDBPath != "Something" {
		t.Errorf("Wrong variable read - %v", cfg.App.BoltDBPath)
	}
	if cfg.App.DataStorePath != "data/export/" {
		t.Errorf("Wrong variable read - %v", cfg.App.DataStorePath)
	}

	if cfg.App.ChangeAcksHeight != 123 {
		t.Errorf("Wrong variable read - %v", cfg.App.ChangeAcksHeight)
	}

}

func TestReadConfig(t *testing.T) {
	fconfig := ReadConfig("")
	if fconfig == nil {
		t.Error("Empty string ReadConfig() should result in a non-nil config")
	}
	fconfig.String()
	GetConfigFilename("")
}

// Check that the home directory is correctly prepended to bare files only
func TestCheckConfigFileName(t *testing.T) {
	checks := map[string]string{
		"C:junk":  "C:junk",
		"~/junk":  "~/junk",
		"\\junk":  "\\junk",
		"./junk":  "./junk",
		"../junk": "../junk",
		"junk":    GetHomeDir() + "/.factom/m2/" + "junk",
	}
	for i, o := range checks {
		name := CheckConfigFileName(i)
		if name != o {
			t.Errorf("CheckConfigFileName(\"%s\")!=\"%s\" instead it it \"%s\"", i, o, name)

		}
	}
}
