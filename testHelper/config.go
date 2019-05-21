package testHelper

import (
	"fmt"
	"os"
	"testing"

	"github.com/FactomProject/factomd/util"
)

var nodeIDs [10]string = [10]string{
	// fnode0 default ID (special bootstrap id)
	`[app]
IdentityChainID                       = 38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9
LocalServerPrivKey                    = 4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d
LocalServerPublicKey                  = cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a
`,
	// fnode01 default ID
	`[app]
IdentityChainID                       = 8888881570f89283f3a516b6e5ed240f43f5ad7cb05132378c4a006abe7c2b93
LocalServerPrivKey                    = 3838383838383135373066383932383366336135313662366535656432343066
LocalServerPublicKey                  = 803b318b23ec15de43db470200c1afb5d1b6156184e247ed035a8f0b6879155b
`,
	// fnode02 default ID
	`[app]
IdentityChainID                       = 8888888da6ed14ec63e623cab6917c66b954b361d530770b3f5f5188f87f1738
LocalServerPrivKey                    = 3838383838383864613665643134656336336536323363616236393137633636
LocalServerPublicKey                  = 11cae6d21e92d9ac0ee83e00f89a3aabde7e3c6f90824339281cfeb93c1377cd
`,
	// fnode03 default ID
	`[app]
IdentityChainID                       = 888888aeaac80d825ac9675cf3a6591916883bd9947e16ab752d39164d80a608
LocalServerPrivKey                    = 3838383838386165616163383064383235616339363735636633613635393139
LocalServerPublicKey                  = 15688e940b854d71411dd8dead29843932fc79c9c99cfb69ca6888b29cd13237
`,
	// fnode04 default ID
	`[app]
IdentityChainID                       = 888888f0b7e308974afc34b2c7f703f25ed2699cb05f818e84e8745644896c55
LocalServerPrivKey                    = 3838383838386630623765333038393734616663333462326337663730336632
LocalServerPublicKey                  = 67bb9fba9c9bab4cc532d9684001ae8bdb70ece551414ff25521d3647370f1c6
`,
	// fnode05 default ID
	`[app]
IdentityChainID                       = 888888d2bc4ed232378c59a85e6c462bcc5495146f3a931a3a1ca42e3397f475
LocalServerPrivKey                    = 3838383838386432626334656432333233373863353961383565366334363262
LocalServerPublicKey                  = d4013c2379a725741534b8f636ada753274722aefa44b91963a104eb9c766b48
`,
	// fnode06 default ID
	`[app]
IdentityChainID                       = 88888867ee42e8b221343da237e08c0b35f50585854c5c05380837da5d55a098
LocalServerPrivKey                    = 3838383838383637656534326538623232313334336461323337653038633062
LocalServerPublicKey                  = 4fb6de25a67608a66c221191f216e0613b21665dc056024f1b4a3cb0b818880a
`,
	// fnode07 default ID
	`[app]
IdentityChainID                       = 888888a5b59731c10c1867474ce26935336ca0269f75a43a903fa4cfeb1aaa98
LocalServerPrivKey                    = 3838383838386135623539373331633130633138363734373463653236393335
LocalServerPublicKey                  = 63ac650e55149eedd01c4df5f74ea74682c6f82a85bedf26adf8b0406a2488bc
`,
	// fnode08 default ID
	`[app]
IdentityChainID                       = 8888887f03e531e68922a71a15bdda9d0430cb5aaaf7ab9f338ba7b5c82d240b
LocalServerPrivKey                    = 3838383838383766303365353331653638393232613731613135626464613964
LocalServerPublicKey                  = e4ab02eb263fad36e2768cf0cb9b50ebcaf779c37b27fef81a24cbb9b1f98424
`,
	// fnode09 default ID
	`[app]
IdentityChainID                       = 888888c0bc99166c1419f86911833a0a1c0b491e79037eeb917ceeabe38232cd
LocalServerPrivKey                    = 3838383838386330626339393136366331343139663836393131383333613061
LocalServerPublicKey                  = 7eef4c8fac8907ad4f34a27c612a417344eb3c2fc1ec9b840693a2b4f90f0204
`}

// Write an identity to a config file for an Fnode, optionally appending extra config data
func WriteConfigFile(identityNumber int, fnode int, extra string, t *testing.T) {
	var simConfigPath string
	var configfile string

	if fnode == 0 {
		simConfigPath = util.GetHomeDir() + "/.factom/m2"
		configfile = fmt.Sprintf("%s/factomd.conf", simConfigPath)
	} else {
		simConfigPath = util.GetHomeDir() + "/.factom/m2/simConfig"
		configfile = fmt.Sprintf("%s/factomd%03d.conf", simConfigPath, fnode)
	}
	if _, err := os.Stat(simConfigPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Creating directory"+simConfigPath+"\n")
		os.MkdirAll(simConfigPath, 0775)
	}
	fmt.Fprintf(os.Stderr, "Create configfile %s\n", configfile)
	f, err := os.Create(configfile)
	if err == nil {
		_, err = f.WriteString(nodeIDs[identityNumber] + extra)
	}
	if err != nil {
		t.Fatal(err)
	}
	return
}
