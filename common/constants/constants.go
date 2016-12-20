// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package constants

import (
	"time"
)

// Messages
const (
	EOM_MSG                       byte = iota // 0
	ACK_MSG                                   // 1
	FED_SERVER_FAULT_MSG                      // 2
	AUDIT_SERVER_FAULT_MSG                    // 3
	FULL_SERVER_FAULT_MSG                     // 4
	COMMIT_CHAIN_MSG                          // 5
	COMMIT_ENTRY_MSG                          // 6
	DIRECTORY_BLOCK_SIGNATURE_MSG             // 7
	EOM_TIMEOUT_MSG                           // 8
	FACTOID_TRANSACTION_MSG                   // 9
	HEARTBEAT_MSG                             // 10
	INVALID_ACK_MSG                           // 11
	INVALID_DIRECTORY_BLOCK_MSG               // 12

	REVEAL_ENTRY_MSG      // 13
	REQUEST_BLOCK_MSG     // 14
	SIGNATURE_TIMEOUT_MSG // 15
	MISSING_MSG           // 16
	MISSING_DATA          // 17
	DATA_RESPONSE         // 18
	MISSING_MSG_RESPONSE  //19

	DBSTATE_MSG          // 20
	DBSTATE_MISSING_MSG  // 21
	ADDSERVER_MSG        // 22
	CHANGESERVER_KEY_MSG // 23
	REMOVESERVER_MSG     // 24

	BOUNCE_MSG      // 25	test message
	BOUNCEREPLY_MSG // 26	test message

	MISSING_ENTRY_BLOCKS //27
	ENTRY_BLOCK_RESPONSE //28
)

const NUM_MESSAGES = 29

const (
	// Replay
	INTERNAL_REPLAY = 1
	NETWORK_REPLAY  = 2
	TIME_TEST       = 4 // Checks the time_stamp;  Don't put actual hashes into the map with this.
	REVEAL_REPLAY   = 8 // Checks for Reveal Entry Replays ... No duplicate Entries within our 4 hours!

	ADDRESS_LENGTH = 32 // Length of an Address or a Hash or Public Key
	// length of a Private Key
	SIGNATURE_LENGTH     = 64    // Length of a signature
	MAX_TRANSACTION_SIZE = 10240 // 10K like everything else?
	// Not sure if we need a minimum amount.  Set at 1 Factoshi

	// Database
	//==================
	// Limit on size of keys, since Maps in Go can't handle variable length keys.

	// Wallet
	//==================
	// Holds the root seeds for address generation
	// Holds the latest generated seed for each root seed.

	// Block
	//==================
	MARKER                  = 0x00                       // Byte used to mark minute boundries in Factoid blocks
	TRANSACTION_PRIOR_LIMIT = int64(12 * 60 * 60 * 1000) // Transactions prior to 12hrs before a block are invalid
	TRANSACTION_POST_LIMIT  = int64(12 * 60 * 60 * 1000) // Transactions after 12hrs following a block are invalid

	//Entry Credit Blocks (For now, everyone gets the same cap)
	EC_CAP = 5 //Number of ECBlocks we start with.
	//Administrative Block Cap for AB messages

	//Limits and Sizes
	//==================
	//Maximum size for Entry External IDs and the Data
	HASH_LENGTH = int(32) //Length of a Hash
	//Length of a signature
	//Prphan mem pool size
	//Transaction mem pool size
	//Block mem bool size
	//MY Process List size

	//Max number of entry credits per entry
	//Max number of entry credits per chain

	COMMIT_TIME_WINDOW = time.Duration(12) //Time windows for commit chain and commit entry +/- 12 hours

	//NETWORK constants
	//==================
	VERSION_0               = byte(0)
	FACTOMD_VERSION         = 4000000
	MAIN_NETWORK_ID  uint32 = 0xFA92E5A2
	TEST_NETWORK_ID  uint32 = 0xFA92E5A3
	LOCAL_NETWORK_ID uint32 = 0xFA92E5A4
	MaxBlocksPerMsg         = 500
)

const (
	// NETWORKS:
	NETWORK_MAIN   int = iota // 0
	NETWORK_TEST              // 1
	NETWORK_LOCAL             // 2
	NETWORK_CUSTOM            // 3
)

// Slices and arrays that should not ever be modified:
//===================================================
// Used as a key in the wallet to find the current seed value.
var CURRENT_SEED = [32]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

// Entry Credit Chain
var EC_CHAINID = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0c}

// Directory Chain
var D_CHAINID = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0d}

// Directory Chain
var ADMIN_CHAINID = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0a}

// Factoid chain
var FACTOID_CHAINID = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x0f}

// Zero Hash
var ZERO_HASH = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var ZERO = []byte{0}

//---------------------------------------------------------------
// Types of entries (transactions) for Admin Block
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#adminid-bytes
//---------------------------------------------------------------
const (
	TYPE_MINUTE_NUM         uint8 = iota // 0
	TYPE_DB_SIGNATURE                    // 1
	TYPE_REVEAL_MATRYOSHKA               // 2
	TYPE_ADD_MATRYOSHKA                  // 3
	TYPE_ADD_SERVER_COUNT                // 4
	TYPE_ADD_FED_SERVER                  // 5
	TYPE_ADD_AUDIT_SERVER                // 6
	TYPE_REMOVE_FED_SERVER               // 7
	TYPE_ADD_FED_SERVER_KEY              // 8
	TYPE_ADD_BTC_ANCHOR_KEY              // 9
	TYPE_SERVER_FAULT
)

//---------------------------------------------------------------------
// Identity Status Types
//---------------------------------------------------------------------
const (
	IDENTITY_UNASSIGNED               int = iota // 0
	IDENTITY_FEDERATED_SERVER                    // 1
	IDENTITY_AUDIT_SERVER                        // 2
	IDENTITY_FULL                                // 3
	IDENTITY_PENDING_FEDERATED_SERVER            // 4
	IDENTITY_PENDING_AUDIT_SERVER                // 5
	IDENTITY_PENDING_FULL                        // 6
	IDENTITY_SKELETON                            // 7 - Skeleton Identity
)

//---------------------------------------------------------------------
// Checkpoints Directory Block KeyMR
//---------------------------------------------------------------------
var CheckPoints = map[uint32]string{
	2:     "5328d4bbe7ea6efc31cf7bfc45192378454cf4e1908c56a35e6a64456a691751",
	10:    "3a5ec711a1dc1c6e463b0c0344560f830eb0b56e42def141cb423b0d8487a1dc",
	48:    "471ef865fecf2b1a98b2e1f87434cc65a6672cc9c8fe19ab2471112431f54b36",
	100:   "cde346e7ed87957edfd68c432c984f35596f29c7d23de6f279351cddecd5dc66",
	150:   "d029c37c0679bfc1cdf1096237f320ae6535def5f64aeffc5105554013aa9e23",
	198:   "534a4188b92155c55f9626bbf4b02721468f6237e467f9c550c03eaafc913003",
	200:   "d13472838f0156a8773d78af137ca507c91caf7bf3b73124d6b09ebb0a98e4d9",
	248:   "37d3801c4079012c989bef9d35cf608773855197acd63f5a253bcfebf74b87d8",
	300:   "4757d31b255e789435c807cb76f5de6cd6590a39a1e5bcfa576d0290eb52dd34",
	348:   "86d2159871316f4868a81586f44b742d5fbe99e2053237ff6575df67657639af",
	400:   "a3c4336ff44989717664233892edfc018c579d868699d9db29f188dbae3a1f3f",
	448:   "abc0048b027735c87761b3b99d69e12e1b6434ff3f2c090b23a9738d440fa9b2",
	480:   "ed0da6e9879495d15425f693c2fb120192298aeaa92772fd9f963b7eb69b1bb6",
	500:   "2978233e69cf207a92bac162598a0398c408caecec7092151db5d044587af5d6",
	1000:  "cd45e38f53c090a03513f0c67afb93c774a064a5614a772cd079f31b3db4d011",
	2000:  "0fae4e8749045bcec480a47019ab2423ac8339d33447cc1f7978395a841b6f55",
	3000:  "599c1e4527cf5880210d21f7aa1063aea68dd1a985d65ba037c57acc433e867e",
	4000:  "3163946232e9e8ec22b21a9db1373c172ebf7a7993dc54c1a0f41f4251e8d7f5",
	5000:  "ca02b78949b80427ddecbbf266d0c18b5dbbbfd840e5d505d64147fa109bf29e",
	6000:  "d1975eb7bd7e0002f7f4d77469a95b466340556ae0461135d7f9469b9eec173e",
	7000:  "91a523f521e910a870c64c155076ceb203b210d009d34e50cc441194ee621de8",
	8000:  "4e2d73d19959240c491df3edc03d02f6a0a2a05f7b75e1b4fe7f299636f073c7",
	9000:  "19da7a9a36dc146c740ab4fc4bbf53b25441fdd8925eda29a8b870cda81a1bb5",
	10000: "3670a63eb8051b925213a4a350e8d37d87e43da8a577a609d7fd30629b73a3aa",
	11000: "53b0884fc5bb9de48db83b5e66d4ca1cb1d29dcb15e865903c37f9137dfe2cc8",
	12000: "c4045aaf92e71ae3c25135022fb2d777164a6530fdd279ea6d5c0d383e87d60d",
	13000: "cff9894749008b42874e6bea9da33d6ba0f7ff85405a10f3b980a900d9618104",
	14000: "491809dc5a07ae895a9aca8634113267a4e38be11bb980012a244cd6559e8308",
	15000: "4358041d6773351dd0a42a8d16778c6544b1196a03c6c41645340cd076a29b6b",
	16000: "1c5a3ea4233871c564b0cae1c01201bafa3f84d3477532bb5318b72fdbc51804",
	17000: "7af7d416d96bbdd0acbeada4b3a70d2e400ec11706db742b6c7c8f60adb35b49",
	18000: "b5837c846cc314c9cedde7a0f0633b9d4f278a867a5808de1bf9da1a6c06e795",
	19000: "8bf15f172bd03f13db2937c28fd333e867faa39a55c09f095f84be6658ff8cea",
	20000: "623f18fb113dca850b78389fab662033f86d65a0efc2f1760f11939f3a8df98a",
	21000: "1bb09820f0c4650d53b3be362242eca5284d43ae74ea88abafc82b5761b1bfce",
	22000: "560312328e848d9e7680370c6e54480389e6ae692ea6ae23a4253bdaf8bace15",
	23000: "91d1fa6b470235cdd334436173620472feecda86cbb122df693066372e411008",
	24000: "23a16b202dc818001457483fcbfab1417aed727aabbb156a4f6ffa82f2736ef2",
	25000: "be8f161e0ffa2e3d50cdbded924ee47e9419bf52b900a4150ef74d9016dfd1c0",
	26000: "7395266728a716c9f4d6f994da64cb9f91a4b4e804cf601dcd371a679d38400a",
	27000: "9b765658a2e5ff36d28f335eacd474d9bfdae8112ae348d2674cb3710fdb9b3b",
	28000: "60f250826550092034003cc1c06c9c33a33bf59733800c6a5e4bac094b62b2d9",
	29000: "0b1e4bdeb1098d590813ce44b9eb256181e5513669cdf93e56f135e2c80cb880",
	30000: "50d9cf6c596a09d0fab37467601a4caa5c7d1b5fd2ee007af25646f6152a392b",
	31000: "b350499dfbb973455385e5b826de77c3e5efcbea0e6388cecf64134416f47c1b",
	32000: "c49cc7d2de2b10feb1c6dafd598485dda0a67fc7584756dc465bacb8bf05090e",
	33000: "eb8e40327d6a60b00e4d23255b29c3b12f8b4093e7f0b266cb9dd25e32ab297d",
	34000: "0879a4866628e0eeab98e479f59ce36d776c17d56849e4d9678182847b5e6689",
	35000: "e630fbd538efcdece0b134ba93b719072d871297e233abda7e79dcc5f3bba9f5",
	36000: "f421b36795edf9b60a74c8f2342e836f5a1693a6cff3e332a886a4783415cacd",
	37000: "9660a52b7e130862d1c043562310f6c98eb5ac9999c827615e73a4693cd75f99",
	38000: "faf8ace8a60a68ff7d2e9b02b145b2958a7ba9c13bf05d55bf12dc8f94bc7c6c",
	39000: "a08aad2aea05be4a4fd4583f068af28601ce9d905c08d07434f2aca1865e6a3e",
	40000: "df11b01490dd5f7e8a849205aa72b56158f3022c33b0075677c747ae4c2cac65",
	41000: "01a9fb685df848f887281decbc2446fd22490305bffbdeb065d937deb34147c3",
	42000: "03497a662826e06645d60c97a8af6e4044654d5513a4c3b8593940973116fb9f",
	43000: "ee21501e47c5307d3bcee7f730fe878f5b331126df688d7c20410b9d75fb6739",
	44000: "6f43842101f04bf6cd62aaa1ba3e98591ec89f31b41c043c3b5ed333e4be7918",
	45000: "87e17c6740c84088d3e9d6d49c51c5999d7a29ae66b73b270661d2ae18b36d11",
	46000: "09d3e1fce45d296f4bc299471ab5edcb2cd3a366c71402230afcf3f69dfe7a9d",
	47000: "acbd841298e84c8edb72db09433d3419964631da21824cdd94c1a1d9bff5ccf3",
	48000: "b5885b4780dd63950a9d69236d57b9d505e1059e0038d26167daaa30e1137120",
	49000: "036fc982d9d534cf31d4d16419964c4cc00c6ad13bf39dc90e0ea5d59dc57d01",
	50000: "5abd0dd2b470c40afd864796a9408fe9a9ed46c360672387cb6a8a09057d3ef3",
	51000: "5b8dfb559c03ab0f73e045c512c92a06d927363cb40b40e42cb746a7884cc8af",
	52000: "6296a645d03e22cfa769ce48ab735e69789f026ca65d20f26c2a8adc2e9bf630",
	53000: "792bce3c65bab4321db09b3f5b017b5496dd217352858215ed058faaa00aefeb",
	54000: "11f44cadaf19eea29dc366a828531e36e8137ea2a5687041cf51e5ad66a7233d",
	55000: "1101b4c1003a393bc17bae9305b103d07ef7b21bc5f170cf41c7e07e8862e6ad",
	56000: "1e893bf343234de2a31192a0e5fff02f785989afd99823c1d37de5af76c5be45",
	57000: "28de098d3c84249e070736b923d026c6f6b432b60efb89678249baf28c6ae9ac",
	58000: "c1345922478345fb4124591583bbf369dcd2b5e7e3aadb0941233285bb89f993",
	59000: "b8d0837c26aacb6355a13aac8d80073f855fdf8c0cf2f528f7daf3258044c8b3",
}
