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
	; --------------- DBType: LDB | Bolt | Map
	DBType                                = "Map"
	LdbPath                               = "ldb"
	BoltDBPath                            = ""
	DataStorePath                         = "data/export/"
	DirectoryBlockInSeconds               = 6
	ExportData                            = true
	ExportDataSubpath                     = "export/"
	; --------------- Network: MAIN | TEST | LOCAL
	Network                               = LOCAL
    PeersFile                             = "~/.factom/peers.json"
	SeedURL                               = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/seed.txt"
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

	[wsapi]
	ApplicationName                       = "Factom/wsapi"
	PortNumber                            = 8088

	; ------------------------------------------------------------------------------
	; logLevel - allowed values are: debug, info, notice, warning, error, critical, alert, emergency and none
	; ConsoleLogLevel - allowed values are: debug, standard
	; ------------------------------------------------------------------------------
	[log]
	logLevel                              = debug
	LogPath                               = "m2factom-d.log"
	ConsoleLogLevel                       = debug

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

}
