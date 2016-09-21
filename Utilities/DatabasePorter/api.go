package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
)

//var server string = "localhost:8088" //Localhost
//var server string = "52.17.183.121:8088" //TestNet
var server string = "52.18.72.212:8088" //MainNet

type DBlockHead struct {
	KeyMR string
}

func GetDBlock(keymr string) (interfaces.IDirectoryBlock, error) {
	raw, err := GetRaw(keymr)
	if err != nil {
		return nil, err
	}
	dblock, err := directoryBlock.UnmarshalDBlock(raw)
	if err != nil {
		return nil, err
	}
	return dblock, nil
}

func GetABlock(keymr string) (interfaces.IAdminBlock, error) {
	raw, err := GetRaw(keymr)
	//fmt.Printf("Raw - %x, %v\n", raw, err)
	if err != nil {
		return nil, err
	}
	switch keymr {
	case "170ef3dc1c079761cf514cbda9689b9f59cbecb50d80cde68ef873f31a7943d8":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a013b755b155d0bcb1131f979d61293549bf9c35dcb31fdf91fa8f89bfdb03f360000d5210000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a1d4ed3fcb3e7020568ff414829b9878c961d83bf4289e91024194e3665e484f92d7cc26eac81ac287b1084f7d138e8abd93ff76c880f8f21172180703f4fe0090001")
		break
	case "20c1897e9633c572bc67adfa83c62b70cfab0045063aed602c18d0adb1a5b411":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a3b3bbb6437124af54aadf8a6b46d0996429ab4ef467e70ae753e2d6d45ba35160000d4f00000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613acfcd165519361e4a4cd6c1fcaa717ee624967b7871f6c1accecb857da57e92f2218b17c08ed594872c7d55ebfbe918c1a1d2ce987b50d7c1c5c4b3491fa91c040001")
		break
	case "01be588d7234ea9b747ff3ba426fbfc27aaf49ea5c0121c2feea52516c65e623":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a2438c907a877b75acff4433ae2efc2636baabbc46d5f6ee4849085a3593329440000d40c0000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a584a7febd4de91ec7ffa6fbd5dccf256da4d136075c5d29e76f091f73b9280e4c94a914099d69a47a70aa0fda067ecd48b7aeecb16cfa1646938b8ca58119f090001")
		break
	case "de66b29ebb1744df698ba2e6dd64862dc96745eff36e2efcd8b3ca79bc930767":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a506fd27224d74cef8a437552efdaaf1d79dc7cfa67f76de56e904ec8588816f00000d4030000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a4d27ace44bf2fc52375d5a95fcd5b401dd0f754616aa8de688c0fc3c3904ae4d36f093e87a5e2f1a1443b6b2bf12f4d8652b6b9215aac9b0b82b436d5788c7040001")
		break
	case "fff49be232d96ce75f69be7b56dfc6f5d4e097efc7b94e971863f85595720689":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a0a763a7d184f44fe0fe31856426f9203e32d6bf24000247b1ba44e8eeb746b1d0000d00b0000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613ab138686153b6331e9251db789e748b741619ae1bb84232f2843ea572a872303366c0f68f7035e547d1ca506022512fd7e693395dcc4e36ed480a95bf985be60e0001")
		break
	case "4dd80ea987406f3134e9ddaac5cadb914bda29a5a55a8ae1156360ebcd6cf5b8":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a3b066846bd2b75f842e0dc88c3c69a0e24f60592fab078ac492c02f146dfd7200000cfa50000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a78d2c33d493ca6e7dba5bf2bc1d635e0b91081ab2521efdee2eb978dc84f0217218ce1adb9518c5ff6279f2242df538b1d9c2c9d1a20bd1f2852c74bb28f3a030001")
		break
	case "f9eea8ea2d5b9a1f3c8009e3f21647370e66368742c65eff342a1019363c22b2":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a6167a1967494e4a0c1fce643d1e80755e891d3392813d37a1d37b0e08b595c260000cf7d0000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a381287c2e20153de68c36b0dcd4e5e8a834dccb12fd7416912d27a4b92ee7f3bf0ec689fdcd9b769797aecc3e8871e3955fb7976979280e320de84a9649a460d0001")
		break
	case "cec4bac15f0be12fd04f56c28666d22cb643cb6bf50621d30505b24cc4d49f07":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a50bd177474eda4d9a4dbb58683fb0a05110ec08567d7c260ba41b5738d77e1520000c93e0000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613ac0bcfd124d4bc3aebf2c24264b52dc1f7c5c8dc38b4ef4db15a24fbe567dfed3c63fe7d548779c13089d963e814ff4a07ea01a9244734edc21ca47dfd38cd3000001")
		break
	case "11aabecc5b3909656aa306d214aa7478c5e757c0661da7ff2d19fa95d0ef227e":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a77606670fc8343468bd3b8bf3bf850d6cb1b1b4dc830dec6259d9d57e5786cc20000c7ab0000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a3288833ff94d8dd714110146333e43241bd4ac3478d47e28dcf44ae97f05449a510ed213d010aab6632a2726390fd6bd8c79b8ee4a637538e36c61f902bf0c0b0001")
		break
	case "3f8bff74c37e8aff95f773d60971ccd39ffc3ede3ee5fc58c44a9fb896602770":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a6876cbc0e9d1d643c9e683064fde0bb3aa7ea4ec840b96d3427136ff6776ff220000c6940000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a1f57817510bb2ab23f62d20d18c5b3e58cefc131e616130c0e4c522f4eae37b586c8a164a4ef6571849ae977af222ffd5aa9a6bd236d9bd90a647e94e8b8d4070001")
		break
	case "0377d2b76c436aa92820af1276f525c441fa38f00477fe898f9fc0917dd8cffb":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000afb2474311b0027246e455a7c3fb38eca662596d1ab07d88ea86c3ba5ffaf1bb30000c3470000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a73267d71c143d993415d30eb888fea085e74efe0c6bfabbe38c9b0115de399602f69173ab0b0de223ec3a416d5637fa53f6e328ea9b0aeb14ae8de6bcef2ae0d0001")
		break
	case "b63aa794c6e409835e35fac507cc16f06e81ccbce74b8f2ca3280a43a65b9089":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000af1d9e71239a876c3b312bb0103a7742ba8f5fb78d28909cce12c0ff42e4dd7e50000c3420000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a3094031f567de0a6d0812b88ca279935810942946ff00c673f0bcbf5ca9f1c338148dbcbf4384057041b74b70d3135ca54add8507adc04c4760484c003e219010001")
		break
	case "5fc677ac4ee0e7d6deefe27acbea90d5fcef9003549ee0875fc0521223ae8b15":
		raw, _ = hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a79a174497a11ea60bf18090434752a57a9b0451f1bd3822d9c03523dea3637380000c2ea0000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a3fa6e4ebd1f2d7af78686f46c6f8a173d053db60bd34563cd4a7135ed1a29428c5f26aa15defeb99d319c7b21d308b9144ba3288b45606036ca20ac46332cc0c0001")
		break
	}

	block, err := adminBlock.UnmarshalABlock(raw)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func GetECBlock(keymr string) (interfaces.IEntryCreditBlock, error) {
	raw, err := GetRaw(keymr)
	if err != nil {
		return nil, err
	}
	block, err := entryCreditBlock.UnmarshalECBlock(raw)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func GetFBlock(keymr string) (interfaces.IFBlock, error) {
	raw, err := GetRaw(keymr)
	if err != nil {
		return nil, err
	}
	block, err := factoid.UnmarshalFBlock(raw)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func GetEBlock(keymr string) (interfaces.IEntryBlock, error) {
	raw, err := GetRaw(keymr)
	if err != nil {
		return nil, err
	}
	block, err := entryBlock.UnmarshalEBlock(raw)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func GetEntry(hash string) (interfaces.IEBEntry, error) {
	raw, err := GetRaw(hash)
	if err != nil {
		fmt.Printf("got error %s\n", err)
		fmt.Printf("called getraw with %s\n", hash)
		fmt.Printf("got result %s\n", raw)

		return nil, err
	}
	entry, err := entryBlock.UnmarshalEntry(raw)
	for err != nil { //just keep trying until it doesn't give an error
		fmt.Printf("got error %s\n", err)
		fmt.Printf("called entryBlock.UnmarshalEntry with %s\n", raw)
		fmt.Printf("got result %s\n", entry)
		//if we get an error like EOF, get the thing again after a short wait
		time.Sleep(20000 * time.Millisecond)
		raw, err = GetRaw(hash)
		if err != nil {
			return nil, err
		}
		entry, err = entryBlock.UnmarshalEntry(raw)
	}
	return entry, nil
}

func GetDBlockHead() (string, error) {
	//return "3a5ec711a1dc1c6e463b0c0344560f830eb0b56e42def141cb423b0d8487a1dc", nil //10
	//return "cde346e7ed87957edfd68c432c984f35596f29c7d23de6f279351cddecd5dc66", nil //100
	//return "d13472838f0156a8773d78af137ca507c91caf7bf3b73124d6b09ebb0a98e4d9", nil //200

	resp, err := http.Get(
		fmt.Sprintf("http://%s/v1/directory-block-head/", server))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf(string(body))
	}

	d := new(DBlockHead)
	json.Unmarshal(body, d)

	return d.KeyMR, nil
}

type Data struct {
	Data string
}

func GetRaw(keymr string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/v1/get-raw-data/%s", server, keymr))
	for err != nil {
		//if the http code gave an error, give a little time and try again before panicking.
		fmt.Printf("got error %s, waiting 20 seconds\n", err)
		time.Sleep(20000 * time.Millisecond)
		resp, err = http.Get(fmt.Sprintf("http://%s/v1/get-raw-data/%s", server, keymr))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	for err != nil {
		//if the io reader code gave an error, give a little time and try again before panicking.
		fmt.Printf("got error %s, waiting 20 seconds\n", err)
		time.Sleep(20000 * time.Millisecond)
		body, err = ioutil.ReadAll(resp.Body)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(string(body))
	}

	d := new(Data)
	if err := json.Unmarshal(body, d); err != nil {
		return nil, err
	}

	raw, err := hex.DecodeString(d.Data)
	if err != nil {
		return nil, err
	}

	return raw, nil
}
