package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type HardGrant struct {
	DBh     uint32
	Amount  uint64
	Address interfaces.IAddress
}

// Return the Hard Coded Grants. Buried in an func so other code cannot easily Address the array and change it
func GetHardCodedGrants() []HardGrant {
	var hardcodegrants []HardGrant

	switch globals.Params.NetworkName {
	case "LOCAL":
		hardcodegrants = []HardGrant{
			// waiting for "real-ish" data from brian
			HardGrant{41, 2, validateAddress("FA3oajkmHMfqkNMMShmqpwDThzMCuVrSsBwiXM2kYFVRz3MzxNAJ")}, // Pay Clay 2
			HardGrant{51, 4, validateAddress("FA3Ga2XcaheS5NgQ3q22gBpLgE6tXmPu1GhjdU2FsdN2QPMzKJET")}, // Pay Bob 4
			HardGrant{41, 3, validateAddress("FA3Ga2XcaheS5NgQ3q22gBpLgE6tXmPu1GhjdU2FsdN2QPMzKJET")}, // Pay Bob 3

			// Note to future grant implementers.  To test the grants that you have coded up on mainnet before deployment on your local machine use this procedure.
			// - Code all the grants and add them to the MAIN section. Use the correct activation height, where Height % 25 = 1
			// - Next copy the new grants to this section, but substituing the activation height for one closer to zero.  the Height should be % 10 = 1
			//   If the grants are activating on multiple different block heights, then use different heights in this section too.
			// - Clear out the .factom/m2/local-database folder to start with a fresh blockchain
			// - Compile and run like this: factomd -network=LOCAL
			// - Open the control panel and wait until after the activation block has been signed.
			// - Open the Admin block at that height (in this example 11) and make sure the coinbase constructor looks correct.  Repeat this for all the activation heights
			// - Wait 10 more blocks and then open the factoid block which implements the grants and doublecheck that the addresses look correct.  Check that the total output balance looks right
			// - To test the ability of the transaction to pass over DBstates do this:
			// - Stop factomd and restart it with this command: factomd -network=LOCAL -count=2
			// - Switch to the second simnode by pressing 1 <enter> in the simulator console
			// - Check you are on the second node by pressing s <enter> to print out the summary.  It should show "1 f FNode01" to indicate focus is on the new simnode.  This means you are viewing that control panel now.
			// - Refersh the control panel and make sure that it downloads the blockchain and is keeping up with the first simnode and is not stalled on either of the grant blocks.

			// - Note, if set to 11, then there may be an extra payout at block 151
			//   the TESTNET_COINBASE_PERIOD has a activation height of 25 for local net instead of maxint so after block 25 the
			//   COINBASE_DECLARATION is changed to 140 so at height 151 we look back 140 blocks and get the grants at height 11
			//   and pay them out a second time. Perhaps we should handle that differently somehow on any future activation
			//   heights that change the COINBASE constants. The bug is TESTNET_COINBASE_PERIOD should never activate for local
			//   networks since it is intended to only apply to test nets.

			// Copy (and replace) the new grants to be tested here:
			// Centis BV total: 1200 FCT (300 FCT * 2 months) + (600 FCT * 1 month) = 1200 FCT lowered upon request
			HardGrant{11, 1200e8, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},
			// The 42nd Factoid total: 1800 FCT
			HardGrant{11, 1800e8, validateAddress("FA3AEL2H9XZy3n199USs2poCEJBkK1Egy6JXhLehfLJjUYMKh1zS")},
			// Factom, Inc. total: 1800 FCT
			HardGrant{11, 1800e8, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},
			// Canonical Ledgers total: 1800 FCT
			HardGrant{11, 1800e8, validateAddress("FA2PEXgRiPd14NzUP47XfVTgEnvjtLSebBZvnM8gM7cJAMuqWs89")},
			// DBGrow total: 1800 FCT
			HardGrant{11, 1800e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},
		}
	case "CUSTOM":
		hardcodegrants = []HardGrant{}

	case "MAIN":
		// Note to future grant implementers.  On mainnet, grants should be on blocks divisible by 25 + 1.  See function CheckGrants().
		hardcodegrants = []HardGrant{
			// Initial grants approved June 9, 2018 https://drive.google.com/drive/folders/1e_xmKgJb375EcAwwkz2d3mdYs0PVVih3
			// https://factomize.com/forums/threads/implementing-the-initial-grants.428/
			// USD/FCT rate calculated at $8.6218 using 7 day EMA from August 1, 2018 https://drive.google.com/drive/folders/1GgAtsTlZEMD77eUvQW3TSGULxIT0Ld03

			/// activation block height of 152751 is expected on Monday, August 6 around 7pm UTC

			// USD denominated grants
			// Legal Review Grant:
			// $200,000 total = 23197 FCT split over 4 addresses
			// part 1/4 8065 FCT
			HardGrant{152751, 806500000000, validateAddress("FA2eNHCf6Sh8aPJrtiparZcxKHWENbvAVvm18Yw9tXyqHyVdxz6E")},
			// part 2/4 8730 FCT
			HardGrant{152751, 873000000000, validateAddress("FA2Ls3yMt9gy8MxxxzvuBj4Zt6wjt15kEW3xoTKi1Npe4LQ1idDw")},
			// part 3/4 2328 FCT
			HardGrant{152751, 232800000000, validateAddress("FA2apt71pNu4dfr7o1zFiCdKUJQvCYfnjXjiUveV8VJNJ9QmXg93")},
			// part 4/4 4074 FCT
			HardGrant{152751, 407400000000, validateAddress("FA2LpYCTSxWssXghqtSkpsedB9zozPy9psMKfbJhcCV1qKLp4iZy")},

			// Voting System Grant:
			// Sent to 4 addresses for 4 parties a total of 26374 FCT
			// TFA $154291.2  ($105091.2 + $43200 + $6000) = 17895 FCT
			// part 1/4 17895 FCT
			HardGrant{152751, 1789500000000, validateAddress("FA3dxxVGuL3oPCBLoxmTPXrse3dyS1VyqTttGWJpcCPjnihFFqT5")},
			// Factomatic $45400 = 5266 FCT
			// part 2/4 5266 FCT
			HardGrant{152751, 526600000000, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},
			// LUCIAP $24200 = 2807 FCT
			// part 3/4 2807 FCT
			HardGrant{152751, 280700000000, validateAddress("FA1yjNL71ddoQPMcgJpBK4TcuyRH8xpHSZWMLJ7DbkdjhHBVZhEm")},
			// Factoshi $3500 = 406 FCT
			// part 4/4 406 FCT
			HardGrant{152751, 40600000000, validateAddress("FA2EPEXRPLc6HE3py95f2bsouFTj7tDq9gd5RVdv1RVAUiUTXyJm")},

			// FCT denominated grants

			// Java Enterprise Client Library Grant:
			// Blockchain Innovation Foundation 1200 FCT
			HardGrant{152751, 120000000000, validateAddress("FA3YVoaN2D8xNitQ6BNhUDW6jH73MdYKyokJqs9LPJM8cuusM7fo")},

			// Oracle Master Grant:
			// Factom, Inc.  300 FCT
			HardGrant{152751, 30000000000, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},

			// Anchor Master grant:
			// Factom, Inc.  One time 600 FCT plus 220 per month from June 9 -> Sep 9, 2018. Total of 1260 FCT
			HardGrant{152751, 126000000000, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},

			// Protocol Development Grant:
			// Factom, Inc.  30000 per month from June 9 -> Sep 9, 2018, Total of 90000 FCT
			HardGrant{155501, 9000000000000, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")}, // c

			// Guide Payments (past) April 7- July 7
			// April 2018.  April 7 - 30 = 160 FCT for each of the 5 guides
			// May 2018. 200 FCT for 3 guides.  Dchapman 70.96 FCT. Canonical Ledgers & DBGrow 6.45 FCT
			// June 2018. pay rate was 200 FCT/month June1-7 and 600 FCT/month June 8-30  40 FCT for each guide for July 1-7.
			// From June 7 - July 7 each guide got 600 FCT

			// Dchapman
			// 160 + 70.96 = 230.96 rounded to 231
			HardGrant{152751, 23100000000, validateAddress("FA22C7H844H5TrWNyXerqp6ZAvdSxYEEUNZxexuZbLnBvcxTobit")},

			// Centis BV total: 160 + 200 + 40 + 600 = 1000 FCT
			HardGrant{152751, 100000000000, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},

			// The 42nd Factoid total: 160 + 200 + 40 + 600 = 1000 FCT
			HardGrant{152751, 100000000000, validateAddress("FA3QzobSZzMMYY1iL8SPWLzRsRggm9cKRg8SXkHxEhaQjqSSLY1o")},

			// Factom, Inc. Factoid total: 160 + 200 + 40 + 600 = 1000 FCT
			HardGrant{152751, 100000000000, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// Matt Osborne/12 Lantern Solutions total: 160 + 200 = 360 FCT
			HardGrant{152751, 36000000000, validateAddress("FA3dug63wat1WnaLSHB3vZw1dbsqTzmgaVqpm727UYKit4sdHgQJ")},

			// Canonical Ledgers total: 6.45 + 40 + 600 = 646.45 rounded to 646 FCT
			HardGrant{152751, 64600000000, validateAddress("FA2PEXgRiPd14NzUP47XfVTgEnvjtLSebBZvnM8gM7cJAMuqWs89")},

			// DBGrow total: 6.45 + 40 + 600 = 646.45 rounded to 646 FCT
			HardGrant{152751, 64600000000, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// Guide Payments (anticipated) July 7 - Sep 7 2018
			// Centis BV total: 600 + 600 = 1200 FCT
			HardGrant{158001, 120000000000, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},

			// The 42nd Factoid total: 600 + 600 = 1200 FCT
			HardGrant{158001, 120000000000, validateAddress("FA3QzobSZzMMYY1iL8SPWLzRsRggm9cKRg8SXkHxEhaQjqSSLY1o")},

			// Factom, Inc. Factoid total: 600 + 600 = 1200 FCT
			HardGrant{158001, 120000000000, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// Canonical Ledgers total: 600 + 600 = 1200 FCT
			HardGrant{158001, 120000000000, validateAddress("FA2PEXgRiPd14NzUP47XfVTgEnvjtLSebBZvnM8gM7cJAMuqWs89")},

			// DBGrow total: 600 + 600 = 1200 FCT
			HardGrant{158001, 120000000000, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// ****Grant Round 2****

			// Governance Grants

			// Guide Payments September 7 - December 7
			// Per the approved FACTOM-GRANT-GUIDES-002, payment is split up into two (2) 900 FCT payments for each
			// Guide. The first payment is to take place at the same time as other grants (approx 2018-11-28T15:00 UTC)
			// and the second to happen at or around 2018-12-7T17:00 UTC.

			// Payment 1 -- Immediate

			// Centis BV total: 900 FCT
			HardGrant{168576, 900e8, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},

			// The 42nd Factoid total: 900 FCT
			HardGrant{168576, 900e8, validateAddress("FA3AEL2H9XZy3n199USs2poCEJBkK1Egy6JXhLehfLJjUYMKh1zS")},

			// Factom, Inc. total: 900 FCT
			HardGrant{168576, 900e8, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// Canonical Ledgers total: 900 FCT
			HardGrant{168576, 900e8, validateAddress("FA2PEXgRiPd14NzUP47XfVTgEnvjtLSebBZvnM8gM7cJAMuqWs89")},

			// DBGrow total: 900 FCT
			HardGrant{168576, 900e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// Payment 2 -- December 7th, 2018

			// Centis BV total: 900 FCT
			HardGrant{169901, 900e8, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},

			// The 42nd Factoid total: 900 FCT
			HardGrant{169901, 900e8, validateAddress("FA3AEL2H9XZy3n199USs2poCEJBkK1Egy6JXhLehfLJjUYMKh1zS")},

			// Factom, Inc. total: 900 FCT
			HardGrant{169901, 900e8, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// Canonical Ledgers total: 900 FCT
			HardGrant{169901, 900e8, validateAddress("FA2PEXgRiPd14NzUP47XfVTgEnvjtLSebBZvnM8gM7cJAMuqWs89")},

			// DBGrow total: 900 FCT
			HardGrant{169901, 900e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// Factom, Inc 004 -- Oracle Master
			// 300 FCT per month for three months + 600 FCT that was failed to be issued last grant round due to
			// a clerical issue = 1500 FCT
			// See: https://github.com/FactomProject/factomd/blob/v6.0.0/state/grants.go#L78, where only 300 FCT
			// are issued instead of the full 900 FCT for the three month period of last grant round
			HardGrant{168576, 1500e8, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},

			// Factom, Inc 005 -- Anchor Master
			// 220 FCT per month for three months = 660 FCT
			HardGrant{168576, 660e8, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},

			// Committee Grants

			// Exchange Committee 001
			// The Exchange Committee was approved to have 5000 FCT to be paid out in installments as required
			// These funds are set aside socially, not to be spent from the grant pool until the Exchange Committee
			// requests them

			//Payment 1, 724 FCT for exchange listing, the Exchange name is under NDA so talk to Sam Vanderwaal
			// with any questions or concerns
			HardGrant{168576, 724e8, validateAddress("FA3YtPXAJehncSQtA8CWgNSWPr5pyeqovGQb99uVdmdeRqKAeg24")},

			// Exchange Committee 002 -- G Suite Service
			HardGrant{168576, 26e8, validateAddress("FA2kd7iAuCrR2GTV39UMaBTphzvQZYVYmvLJYGsjoJRjEGoVNQFd")},

			// Marketing Committee 001
			// 5000 FCT to be paid out immediately
			HardGrant{168576, 5000e8, validateAddress("FA3KhPnfjM8M8cXNvHZGpLMweEzq9irUzVjhRzyXeKayg1Yc1Yzr")},

			// Core Committee 001 -- Bug Bounty
			// Similar to the Exchange Committee, the Core Committee is asking to have 500 FCT set aside to be used
			// at a later date

			// Other Grants

			// Factom, Inc 006 -- Protocol Development

			// Factom, Inc -- 50,000 FCT
			HardGrant{168576, 50000e8, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")},

			// Sponsor 1, Dominic Luxford -- 800 FCT
			HardGrant{168576, 800e8, validateAddress("FA27Y2fEsaBPeFsN87czeZxLsA9fxi3fcy4f4xHXdF58W7TgbaCB")},

			// Sponsor 2a, Nic Robinette -- 298.9 FCT, prorated -- Sept 9 to Oct 12
			HardGrant{168576, 2989e7, validateAddress("FA2TDwdBLoCtEKrWqf7gSDhXzw8F5GLTK2wFLgg3erC1Ms6jBWuk")},

			// Sponsor 2b, Nolan Bauer, replacing Nic Robinette -- 182.6 FCT, prorated -- Nov 19 to Dec 9
			HardGrant{168576, 1826e7, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},

			// Sponsor 3, Factomatic -- 182.6 FCT, prorated -- Nov 19 to Dec 9
			HardGrant{168576, 1826e7, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},

			// BIF 001 -- Factom Core Development
			// 18,500 FCT
			HardGrant{168576, 18500e8, validateAddress("FA2YFZrZkywe1TvSrsWCaZ5RyJ1ZXcN5a2x9WqCRobj5GQawpTNt")},

			// BIF 003 -- Open-source ALfresco integration for Factom
			// 750 FCT
			HardGrant{168576, 750e8, validateAddress("FA2zcJZsucB6Xb6SdDFnACxHeUyY3MYVv1Dadijd5dbB3USYUJjx")},

			//  FACTOM-GRANT-BEDROCK-CRYPTOVIKINGS-DEFACTO-TFA-001 -- Community Courtesy Node System

			// Bedrock Solutions -- 187.5 FCT
			HardGrant{168576, 1875e7, validateAddress("FA2reVN9ARd4scQcVBsepHwk1ex4hVUavLFfB6HwCQcH2TpvJSse")},

			// CryptoVikings -- 187.5 FCT
			HardGrant{168576, 1875e7, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},

			// De Facto -- 187.5 + 125 =  312.5 FCT
			HardGrant{168576, 1875e7, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			HardGrant{168576, 125e8, validateAddress("FA2rrwFVvkFYwyGFHVBMwRqTpycuZiagrQdcbPWzuoEwJQxjDwi3")},

			// TFA -- 187.5 FCT
			HardGrant{168576, 1875e7, validateAddress("FA2LV4s7LKA9BTgWaJNvcr9Yq8rpiH2XD3vEPY3nwSiNSrnRgkpK")},

			// BIF-Factomatic 001 -- Decentralized Identifiers (DIDs)

			// BIF -- 4500 FCT
			HardGrant{168576, 4500e8, validateAddress("FA2GqMAxcx8WonfEV8sNfeeQYa4fnNU3AhzDDzbz7FvjTvQ6tBCH")},

			// Factomatic -- 4000 FCT
			HardGrant{168576, 4000e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},

			// DBGrow -- Factom Protocol Website
			// 1000 FCT
			HardGrant{168576, 1000e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// DBGrow Inc 002 -- FAT Protocol

			// DBGrow -- 18,750 FCT
			HardGrant{168576, 18750e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// Canonical Ledgers -- 7,750
			HardGrant{168576, 7750e8, validateAddress("FA2nBeBX75R7ECdhZS61DLpP5apaS32zwSYQ7aRkahwAjy5bryFo")},

			// Luciap -- 4,250 FCT
			HardGrant{168576, 4250e8, validateAddress("FA3WP4zXozVbKoeUENNojLjWCtUEPhfNgbwbZgrftZ7NPAqJWpds")},

			// LayerTech -- 3,750 FCT
			HardGrant{168576, 3750e8, validateAddress("FA2VkNgMwJJuNvU66ycDWSRYpj3VvwVMjHiUDwpoXmywPFjDCy6D")},

			// FACTOM-GRANT-LEDGER-FACTOM-ID-TFA-BEDROCK-R2-002 -- Factom Identity on Ledger Nano S

			// TFA -- 2,933 FCT
			HardGrant{168576, 2933e8, validateAddress("FA2Taf8n2TNzx8DGEmPGb2yrwWccVehGzt3zjAoFZREpbnF2c9YM")},

			// Bedrock -- 1,334 FCT
			HardGrant{168576, 1334e8, validateAddress("FA2reVN9ARd4scQcVBsepHwk1ex4hVUavLFfB6HwCQcH2TpvJSse")},

			// Sponsor, David Chapman -- 223 FCT
			HardGrant{168576, 223e8, validateAddress("FA3YtPXAJehncSQtA8CWgNSWPr5pyeqovGQb99uVdmdeRqKAeg24")},

			// ********************************
			// **** Grant Round 3 (2019-1) ****
			// ********************************

			// Governance Grants

			// Guide Payments December 7 - March 7
			// Per the approved FACTOM-GRANT-GUIDES-003

			// Centis BV total: 1200 FCT (300 FCT * 2 months) + (600 FCT * 1 month) = 1200 FCT lowered upon request
			HardGrant{181001, 1200e8, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},

			// The 42nd Factoid total: 1800 FCT
			HardGrant{181001, 1800e8, validateAddress("FA3AEL2H9XZy3n199USs2poCEJBkK1Egy6JXhLehfLJjUYMKh1zS")},

			// Factom, Inc. total: 1800 FCT
			HardGrant{181001, 1800e8, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// Canonical Ledgers total: 1800 FCT
			HardGrant{181001, 1800e8, validateAddress("FA2PEXgRiPd14NzUP47XfVTgEnvjtLSebBZvnM8gM7cJAMuqWs89")},

			// DBGrow total: 1800 FCT
			HardGrant{181001, 1800e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},
			// --------------------------------------------------------

			// Anchor and Oracle master grants

			// Factom, Inc 007 -- Oracle Master Dec 9 2018 - Mar 9 2019
			// 300 FCT per month for three months = 900 FCT
			HardGrant{181001, 900e8, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},

			// Factom, Inc 010 -- Oracle Master Mar 9 2019 - June 9 2019
			// 300 FCT per month for three months = 900 FCT
			HardGrant{181001, 900e8, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},

			// Factom, Inc 008 -- Anchor Master Dec 9 2018 - Mar 9 2019
			// 220 FCT per month for three months = 660 FCT
			HardGrant{181001, 660e8, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},

			// Factom, Inc 011 -- Anchor Master Mar 9 2019 - June 9 2019
			// 220 FCT per month for three months = 660 FCT
			HardGrant{181001, 660e8, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},
			// --------------------------------------------------------

			// Committee Grants

			// FACTOM-GRANT-FACTOM-MARKETING-COMMITTEE-002
			// Factom Protocol Marketing Committee - Explainer Video Grant
			HardGrant{181001, 3000e8, validateAddress("FA3QcDsGS2pK6LDuqCuA2i1sRVeH59PhdHPeQ6bneMDj7ZAqbsZg")},

			// FACTOM-GRANT-FACTOM-MARKETING-COMMITTEE-003
			// Factom Protocol Marketing Committee - Hackathon Grant
			HardGrant{181001, 3600e8, validateAddress("FA2pdi4o4qJd2y7ygHbyvJTk6oXrPLJvH27zdLZrUkmc9gT6Mpca")},
			// --------------------------------------------------------

			// FACTOM-GRANT-EXCHANGE-COMMITTEE-001
			// Exchange Committee Funding (held back on request in 2018-2 round, paid out during 2019-1 round, see above approved grant proposal)
			HardGrant{181001, 4276e8, validateAddress("FA2feHES9FUQwSDYHeGT8UasUKAPNb91fMDTi7qqpXqHedrKyDwv")},

			// FACTOM-GRANT-EXCHANGE-COMMITTEE-003
			// Exchange Committee Funding (2019-1 round)
			HardGrant{181001, 5000e8, validateAddress("FA2feHES9FUQwSDYHeGT8UasUKAPNb91fMDTi7qqpXqHedrKyDwv")},
			// --------------------------------------------------------

			// Core Development

			// Factom Inc, Protocol Development Continuation 009 Dec 9 2018 - Mar 9 2019
			HardGrant{181001, 27440e8, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")},
			// Sponsor 1, Dominic Luxford -- 600 FCT
			HardGrant{181001, 600e8, validateAddress("FA27Y2fEsaBPeFsN87czeZxLsA9fxi3fcy4f4xHXdF58W7TgbaCB")},
			// Sponsor 2, Nolan Bauer -- 600 FCT
			HardGrant{181001, 600e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},
			// Sponsor 3, Factomatic -- 600 FCT
			HardGrant{181001, 600e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},
			// Transition assistent, Nic Robinette -- 200 FCT
			HardGrant{181001, 200e8, validateAddress("FA2TDwdBLoCtEKrWqf7gSDhXzw8F5GLTK2wFLgg3erC1Ms6jBWuk")},
			// --------------------------------------------------------

			// Factom Inc, Protocol Development 012 Mar 9 2019 - Jun 9 2019
			HardGrant{181001, 35459e8, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")},
			// Sponsor 1, Dominic Luxford -- 600 FCT
			HardGrant{181001, 600e8, validateAddress("FA27Y2fEsaBPeFsN87czeZxLsA9fxi3fcy4f4xHXdF58W7TgbaCB")},
			// Sponsor 2, Nolan Bauer -- 600 FCT
			HardGrant{181001, 600e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},
			// Sponsor 3, Factomatic -- 600 FCT
			HardGrant{181001, 600e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},
			// --------------------------------------------------------

			// FACTOM-GRANT-FACTOMIZE-002
			// Factomize, Core Code Development Grant
			HardGrant{181001, 6000e8, validateAddress("FA3XkRCucFVp2ZMnY5uSkxrzKojirkeY6KpwkJyNZPRJ4LsjmFDp")},
			// --------------------------------------------------------

			// FACTOM-GRANT-LAYERTECH-001
			// LayerTech, Core Code Development Grant
			HardGrant{181001, 5500e8, validateAddress("FA2qGCTMiufU1cStopyx3NbNwG1Sawpo8MM9icvKXouzA6mSsFbA")},
			// --------------------------------------------------------

			// FACTOM-GRANT-BIF-004 BIF
			// BIF, Core Code Development Grant
			HardGrant{181001, 500e8, validateAddress("FA2YFZrZkywe1TvSrsWCaZ5RyJ1ZXcN5a2x9WqCRobj5GQawpTNt")},
			// --------------------------------------------------------

			// Factom Open API

			// FACTOM-GRANT-DEFACTO-001
			// Factom Open API — Grant #1
			// De Facto
			HardGrant{181001, 6980e8, validateAddress("FA2rrwFVvkFYwyGFHVBMwRqTpycuZiagrQdcbPWzuoEwJQxjDwi3")},
			// Jay Cheroske (Bedrock Solutions)
			HardGrant{181001, 400e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},
			// --------------------------------------------------------

			// Factom Open Node

			// FACTOM-GRANT-BEDROCK-CRYPTLOGIC-DEFACTO-TFA-002
			// Factom Open Node (ex. Courtesy Node) Continuity
			// Bedrock Solutions -- 136.25 FCT
			HardGrant{181001, 13625e6, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},
			// De Facto -- 136.25 FCT
			HardGrant{181001, 13625e6, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			// CryptoLogic -- 136.25 FCT
			HardGrant{181001, 13625e6, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},
			// TFA -- 136.25 FCT
			HardGrant{181001, 13625e6, validateAddress("FA2LV4s7LKA9BTgWaJNvcr9Yq8rpiH2XD3vEPY3nwSiNSrnRgkpK")},
			// --------------------------------------------------------

			// BEDROCK-DEFACTO-001
			// Factom Open Node Enhancement -- 374 FCT total
			// Bedrock Solutions -- 172 FCT
			HardGrant{181001, 172e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},
			// De Facto -- 172 FCT
			HardGrant{181001, 172e8, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			// CryptoLogic -- 30 FCT
			HardGrant{181001, 30e8, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},
			// --------------------------------------------------------

			// Blockchain Expo Global 2019

			// FACTOM-GRANT-PRESTIGE_IT-001
			// Prestige IT - Blockchain Expo Global 2019 (London)
			HardGrant{181001, 682e8, validateAddress("FA3iRzBGA78gkkJ88PinKi3wwNfBhyoGExgzYx9btJZqo5or1o5A")},
			// --------------------------------------------------------

			// Marketing Videos

			// FACTOM-GRANT-FACTOMIZE-001
			// Factomize, Marketing videos
			HardGrant{181001, 500e8, validateAddress("FA3XkRCucFVp2ZMnY5uSkxrzKojirkeY6KpwkJyNZPRJ4LsjmFDp")},
			// --------------------------------------------------------

			// FAT protocol

			// FACTOM-GRANT-DBGrow-Luciap-Canonical Ledgers-002
			// FAT Protocol Development Grant II - 12.700 FCT
			// DBGrow -- 5.500 FCT
			HardGrant{181001, 5500e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},
			// Luciap -- 3.200 FCT
			HardGrant{181001, 3200e8, validateAddress("FA3DikVW7pzhMkJXuP9xszf9o3aKrMHqEpPkLee2Nb6WewhupyM8")},
			// Canonical Ledgers -- 4.000 FCT
			HardGrant{181001, 4000e8, validateAddress("FA2nBeBX75R7ECdhZS61DLpP5apaS32zwSYQ7aRkahwAjy5bryFo")},
			// --------------------------------------------------------

			// ********************************
			// **** Grant Round (2019-2) ****
			// ********************************

			// Governance Grants

			// FACTOM-GRANT-GUIDES-005 -- 9000 FCT
			// Guide Payments 2019-03-07 - 2019-06-07

			// The 42nd Factoid total: 1800 FCT
			HardGrant{194126, 1800e8, validateAddress("FA3AEL2H9XZy3n199USs2poCEJBkK1Egy6JXhLehfLJjUYMKh1zS")},

			// Centis BV total: 1800 FCT
			HardGrant{194126, 1800e8, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},

			// Factom Inc total: 1800 FCT
			HardGrant{194126, 1800e8, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// DBGrow Inc total: 1800 FCT
			HardGrant{194126, 1800e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// Canonical Ledgers: 600 FCT
			// (2019-03-07 - 2019-04-07)
			HardGrant{194126, 600e8, validateAddress("FA2PEXgRiPd14NzUP47XfVTgEnvjtLSebBZvnM8gM7cJAMuqWs89")},

			// TRGG3R LLC: 1200 FCT
			// (2019-04-07 - 2019-06-07)
			HardGrant{194126, 1200e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},
			// --------------------------------------------------------

			// Anchor and Oracle master grants

			// Factom-Inc-013 -- 900 FCT
			// Oracle Master -- (2019-06-09 - 2019-09-09)
			HardGrant{194126, 900e8, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},

			// Factom-Inc-014 -- 660 FCT
			// Anchor Master -- (2019-06-09 - 2019-09-09)
			HardGrant{194126, 660e8, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},
			// --------------------------------------------------------

			// Committee Grants

			// The Core Committee has via grant Core-Committee-002 been awarded an additional 500 FCT grant
			// to be paid out in installments upon request from the committee. The current total amount set
			// aside for the Core Committee is 1000 FCT.
			// --------------------------------------------------------

			// Core Development

			// Factom-Inc-015 -- 36494 FCT
			// Protocol Development -- (2019-06-09 - 2019-09-09)
			HardGrant{194126, 34694e8, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")},
			// Sponsor 1, Dominic Luxford -- 300 FCT
			HardGrant{194126, 300e8, validateAddress("FA27Y2fEsaBPeFsN87czeZxLsA9fxi3fcy4f4xHXdF58W7TgbaCB")},
			// Sponsor 2, Nolan Bauer -- 300 FCT
			HardGrant{194126, 300e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},
			// Sponsor 3, Factomatic -- 300 FCT
			HardGrant{194126, 300e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},

			// Factomize-004 -- 5225 FCT
			// Core Development -- (3 months of development)
			HardGrant{194126, 5225e8, validateAddress("FA3XkRCucFVp2ZMnY5uSkxrzKojirkeY6KpwkJyNZPRJ4LsjmFDp")},

			// Layertech-002 -- 3600 FCT
			// Core Development -- (3 months of development)
			HardGrant{194126, 3600e8, validateAddress("FA2qGCTMiufU1cStopyx3NbNwG1Sawpo8MM9icvKXouzA6mSsFbA")},

			// Sphereon-003 -- 6750 FCT
			// Core Development -- (3 months of development)
			HardGrant{194126, 6750e8, validateAddress("FA3j1ngrZAGHuiWPMbZFjHjvE9q4YUA2g5PKvJF63ZK5VHNeYbAJ")},
			// --------------------------------------------------------

			// Factom Open Node

			// Bedrock-CryptoLogic-DeFacto-TFA-003 -- 360 FCT
			// Open Node Hosting -- (2019-06-01 - 2019-08-31)
			// Bedrock Solutions -- 90 FCT
			HardGrant{194126, 90e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},
			// De Facto -- 90 FCT
			HardGrant{194126, 90e8, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			// CryptoLogic -- 90 FCT
			HardGrant{194126, 90e8, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},
			// TFA -- 90 FCT
			HardGrant{194126, 90e8, validateAddress("FA2LV4s7LKA9BTgWaJNvcr9Yq8rpiH2XD3vEPY3nwSiNSrnRgkpK")},

			// Bedrock-Defacto-001/2 -- 926 FCT
			// Open Node Enhancement (Leftover payment)
			// Bedrock Solutions -- 428 FCT
			HardGrant{194126, 428e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},
			// De Facto -- 428 FCT
			HardGrant{194126, 428e8, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			// CryptoLogic - 70 FCT
			HardGrant{194126, 70e8, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},
			// --------------------------------------------------------

			// FAT protocol related grants

			// DBGrow-004 -- 3972 FCT
			// FAT Development -- (3 months of development)
			HardGrant{194126, 3972e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// DBGrow-Factom Inc-001 -- 4333 FCT
			// FAT Smart Contracts
			// DBGrow -- 3611 FCT
			HardGrant{194126, 3611e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},
			// Factom Inc -- 722 FCT
			HardGrant{194126, 722e8, validateAddress("FA35Kd1Ac1aQXEHPxYTR6jNDPuwVoh6APQQGKViceQTedyE7J2vV")},
			// Factomize LLC (Sponsor) -- 200 FCT
			HardGrant{194126, 200e8, validateAddress("FA3XkRCucFVp2ZMnY5uSkxrzKojirkeY6KpwkJyNZPRJ4LsjmFDp")},

			// TFA-002 -- 1560 FCT
			// FAT-integration into TFA-explorer
			HardGrant{194126, 1560e8, validateAddress("FA2TzC4d9s14zyXgZhpqgrkKQmL1Hf18dkAJYD5kU7b6sFJrGGHa")},

			// LUCIAP-001 -- 1876 FCT
			// FAT Wallet Ledger Nano Support
			HardGrant{194126, 1876e8, validateAddress("FA2T1tgVwrHDVpMqHRRz5676x4CHkZqXGGp1CmBarYg5ZWcU85g4")},

			// TFA-001 -- 4860
			// FAT Firmware upgrade for Ledger Nano S and X
			HardGrant{194126, 4860e8, validateAddress("FA3sUHyYThjwJSSnun5jx91obexMJibaUCeGFoZ9S1SBzcY1xPCP")},
			// --------------------------------------------------------

			// Miscellaneous Grants

			// Go-Immutable-001 -- 20000 FCT
			// Comprehensive Market Strategy & Execution
			HardGrant{194126, 20000e8, validateAddress("FA2d94Yx4RwQ2bC7G1yLGQiuBoMHumJrpuZyunNmsyyyYfXACEvS")},

			// Anchorblock-002 -- 260 FCT
			// Daily Digest -- (3.5 months of previous work)
			HardGrant{194126, 260e8, validateAddress("FA2WvyDhuKerk3RyVeSiTUonmcn7RuPRJGFQB6oCfroSys7NW3q2")},

			// Federate-This-001 -- 3295 FCT
			// Off-Blocks -- (2019-06-24 - 2019-07-21)
			HardGrant{194126, 3295e8, validateAddress("FA2tCnVKbLMjnLfj9nJQCvQ2GyyuW64mYep1CzsQaL5WmFV5Vhdw")},

			// Factomize-003 -- 1980 FCT
			// Authomated Grant System
			HardGrant{194126, 1980e8, validateAddress("FA3XkRCucFVp2ZMnY5uSkxrzKojirkeY6KpwkJyNZPRJ4LsjmFDp")},

			// Defacto-001/2 -- 5000 FCT
			// Factom Open API - Sprint 2 (Admin API, Web UI, Callbacks)
			// De Facto -- 4700 FCT
			HardGrant{194126, 4700e8, validateAddress("FA2rrwFVvkFYwyGFHVBMwRqTpycuZiagrQdcbPWzuoEwJQxjDwi3")},
			// Bedrock Solutions -- 300 FCT
			HardGrant{194126, 300e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},

			// Sphereon-004 -- 6750 FCT
			// Integration of DAML with Factom protocol
			HardGrant{194126, 6750e8, validateAddress("FA2QWEguNjfme5EMyfe1Df4smyfPbJPQT2Ud6zmk5nh67tQm6xq2")},

			// Sphereon-005 -- 3699 FCT
			// Factom Badges
			HardGrant{194126, 3699e8, validateAddress("FA3BM4Mp6L1uDHCLEy9yZXoW8MDUeyJ2tsY5pa4xrQb9EJqqx3Bf")},

			// BIF-007 -- 1700 FCT
			// Identity, DID and signing FIPs
			HardGrant{194126, 1700e8, validateAddress("FA37dNuCWPPwAwWTVvUCGyMgoWQ8RV8ptmGqxKWHW3T9JC4eVPDR")},

			// Triall-002 -- 4500 FCT
			// DIDs for Alfresco
			HardGrant{194126, 4500e8, validateAddress("FA2ncB92aLCcHRmHcNko6MSFyqYVhX67imtJmoZQskLYSMyXdFgq")},

			// BIF-Factomatic-003 -- 8500 FCT
			// Verifiable Claims FIP

			// Factomatic -- 7000 FCT
			HardGrant{194126, 7000e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},
			// BIF -- 1500 FCT
			HardGrant{194126, 1500e8, validateAddress("FA3kYLCT6zo3EvLSQTiBfLrQds77q3xFv9YfqLDKFuchs8Lhh7iP")},

			// ********************************
			// **** Grant Round 2019-3 ****
			// ********************************

			// Note: From this grant round forward the text for the PR is auto-generated by the Factom Protocol
			// Forum Grant Voting System. Thus the grants will not be sorted by grant type as before but be
			// listed by the grant score in descending order.

			// --------------------------------------------------------
			// [Factom Inc.-16] Oracle Master -- 900 FCT

			// Factom Inc. total: 900 FCT
			HardGrant{207326, 900e8, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Factom Inc.-17] Anchor Master -- 660 FCT

			// Factom Inc. total: 660 FCT
			HardGrant{207326, 660e8, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [The 42nd Factoid AS-Centis BV-Factom Inc.—DBGrow Inc-TRgg3R LLC-6] Guide Compensation -- 9000 FCT
			// Guide Compensation 2019-07-21 - 2019-08-17

			// Factom Inc. total: 1800 FCT
			HardGrant{207326, 1800e8, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// TRGG3R LLC total: 1800 FCT
			HardGrant{207326, 1800e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},

			// DBGrow Inc total: 1800 FCT
			HardGrant{207326, 1800e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// The 42nd Factoid AS total: 1800 FCT
			HardGrant{207326, 1800e8, validateAddress("FA3AEL2H9XZy3n199USs2poCEJBkK1Egy6JXhLehfLJjUYMKh1zS")},

			// Centis BV total: 1800 FCT
			HardGrant{207326, 1800e8, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Factomize-7] Core and General Development -- 8747 FCT

			// Factomize total: 8747 FCT
			HardGrant{207326, 8747e8, validateAddress("FA3nsSjUy5uSkqMEug8t3VcehZn5w2ciSMpgqFEEsMRwMrHoa9k3")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Sphereon-7] Core development continuation -- 8093 FCT

			// Sphereon total: 8093 FCT
			HardGrant{207326, 8093e8, validateAddress("FA3P6cLv2Ct2P6SStqhhqagFp9u8kpca9XyJHbhB68v1jeZNFpG8")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Bedrock Solutions-De Facto-Crypto Logic-The Factoid Authority-4] Factom Open Node Continuity -- 1126 FCT

			// The Factoid Authority total: 188 FCT
			HardGrant{207326, 188e8, validateAddress("FA2LV4s7LKA9BTgWaJNvcr9Yq8rpiH2XD3vEPY3nwSiNSrnRgkpK")},

			// Crypto Logic total: 188 FCT
			HardGrant{207326, 188e8, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},

			// Bedrock Solutions total: 375 FCT
			HardGrant{207326, 375e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},

			// De Facto total: 375 FCT
			HardGrant{207326, 375e8, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Factom Inc.-18] Protocol Development -- 38940 FCT

			// Nolan Bauer total: 300 FCT
			HardGrant{207326, 300e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},

			// factomatic total: 300 FCT
			HardGrant{207326, 300e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},

			// David Kuiper total: 300 FCT
			HardGrant{207326, 300e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},

			// Factom Inc. total: 38040 FCT
			HardGrant{207326, 38040e8, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Matt York-Luciap Technology Inc.-DBGrow-Canonical Ledgers-1] FAT Smart Contracts 2 - Development -- 13750 FCT

			// David Chapman total: 300 FCT
			HardGrant{207326, 300e8, validateAddress("FA3nsSjUy5uSkqMEug8t3VcehZn5w2ciSMpgqFEEsMRwMrHoa9k3")},

			// Luciap Technology Inc. total: 200 FCT
			HardGrant{207326, 200e8, validateAddress("FA2kWsMnSXTV9RS2yEGkEGP2yUFrawZQVj5Hp8WqPnZUUBFCpiaH")},

			// Matt York total: 200 FCT
			HardGrant{207326, 200e8, validateAddress("FA35Kd1Ac1aQXEHPxYTR6jNDPuwVoh6APQQGKViceQTedyE7J2vV")},

			// Canonical Ledgers total: 5500 FCT
			HardGrant{207326, 5500e8, validateAddress("FA2xccSAfhGm5k4tPaXF9741xkQ52drWjoJodQhpPxDxepdqasMM")},

			// DBGrow total: 7550 FCT
			HardGrant{207326, 7550e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Go Immutable-2] PR & Social Media Management Continuation -- 22500 FCT

			// Go Immutable total: 22500 FCT
			HardGrant{207326, 22500e8, validateAddress("FA2LwHgp7LyoY5SG1vxjLaQ2QtdyepfaqYUTsiTgprtkiP6geCk7")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Factoshi-1] Graff - A GraphQL wrapper for the factomd RPC API -- 2950 FCT

			// Factoshi total: 2950 FCT
			HardGrant{207326, 2950e8, validateAddress("FA2tsEih6kyzNSBtbkzZ31HUGF8QQB7RPBFouBVfYyUZk24QkPr2")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Marketing Committee-2] Rebranding Services -- 10000 FCT

			// Marketing Committee total: 10000 FCT
			HardGrant{207326, 10000e8, validateAddress("FA3i8WbLoRvmWmpRLEBYENLZdEBK9TNG74UWrvHAGWgAyW11iHBu")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [AroundTheBox-1] Develop Free Excel Add-In -- 680 FCT

			// AroundTheBox total: 680 FCT
			HardGrant{207326, 680e8, validateAddress("FA28qK3QjFEupwjmY3qvM1CdNCrik33my8fdmxJwj1ea5iKCnbXF")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// [Kompendium-4] Rust Client Library for the Factom Protocol -- 2400 FCT

			// Kompendium total: 2400 FCT
			HardGrant{207326, 2400e8, validateAddress("FA3KMPNX8AKdY3tjyKMzZ5cAkqUv97d3QqkJeQnVXk6PXSVgArnr")},
			// --------------------------------------------------------

			// ********************************
			// **** Grant Round 2019-04 ****
			// ********************************

			// --------------------------------------------------------
			// Samuel Vanderwaal-Brian Deery-Nolan Bauer-Nic R-The 42nd Factoid AS-Centis BV -- 9000 FCT
			// Guide Compensation 2019-09-07 to 2019-12-07

			// Factom, Inc. total: 1800 FCT
			HardGrant{221126, 1800e8, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// Nolan Bauer total: 1800 FCT
			HardGrant{221126, 1800e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},

			// Nic R total: 1800 FCT
			HardGrant{221126, 1800e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// The 42nd Factoid AS total: 1800 FCT
			HardGrant{221126, 1800e8, validateAddress("FA3AEL2H9XZy3n199USs2poCEJBkK1Egy6JXhLehfLJjUYMKh1zS")},

			// Centis BV total: 1800 FCT
			HardGrant{221126, 1800e8, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Factom Inc. -- 900 FCT
			// Oracle Master Dec 9 2019 - March 9, 2020

			// Factom Inc. total: 900 FCT
			HardGrant{221126, 900e8, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Factom Inc. -- 660 FCT
			// Anchor Master Dec 9 2019 - March 9, 2020

			// Factom Inc. total: 660 FCT
			HardGrant{221126, 660e8, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// David Chapman-Factomize -- 15769 FCT
			// Core and General Development 2019-12-01 - 2020-02-29

			// Factomize total: 15769 FCT
			HardGrant{221126, 15769e8, validateAddress("FA3nsSjUy5uSkqMEug8t3VcehZn5w2ciSMpgqFEEsMRwMrHoa9k3")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// CryptoLogic-The Factoid Authority-Crypto Logic-Bedrock Solutions-De Facto -- 1956 FCT
			// Factom Open Node Continuity 2019-12-01 - 2020-02-29

			// The Factoid Authority total: 326 FCT
			HardGrant{221126, 326e8, validateAddress("FA2LV4s7LKA9BTgWaJNvcr9Yq8rpiH2XD3vEPY3nwSiNSrnRgkpK")},

			// Crypto Logic total: 326 FCT
			HardGrant{221126, 326e8, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},

			// Bedrock Solutions total: 652 FCT
			HardGrant{221126, 652e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},

			// De Facto total: 652 FCT
			HardGrant{221126, 652e8, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Nolan Bauer-factomatic-David Kuiper-Factom Inc. -- 34740 FCT
			// Protocol Development Dec 9 2019 - March 9, 2020

			// Nolan Bauer total: 300 FCT
			HardGrant{221126, 300e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},

			// factomatic total: 300 FCT
			HardGrant{221126, 300e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},

			// David Kuiper total: 300 FCT
			HardGrant{221126, 300e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},

			// Factom Inc. total: 33840 FCT
			HardGrant{221126, 33840e8, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Samuel Vanderwaal-Exchange Committee -- 5000 FCT
			// Legal Fees Reimbursement

			// Exchange Committee total: 5000 FCT
			HardGrant{221126, 5000e8, validateAddress("FA2fVnbYw3Hr1MCJWktLnGDrYiLnW2HwDBDVRpp3Q355jJRDwvCs")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// David Chapman-Factomize -- 12069 FCT
			// ANO Promotion / Demotion System 2019-12-01 - 2020-02-29

			// Factomize total: 12069 FCT
			HardGrant{221126, 12069e8, validateAddress("FA3nsSjUy5uSkqMEug8t3VcehZn5w2ciSMpgqFEEsMRwMrHoa9k3")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// mboender-Sphereon -- 14909 FCT
			// Core Development Continuation 2019-12-02 - 2020-02-28

			// Sphereon total: 14909 FCT
			HardGrant{221126, 14909e8, validateAddress("FA3igxrULYqL5w4oq9vXwWzWURtvejfYh64iPZaeRqZfLpFdkFgD")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Samuel Vanderwaal-Exchange Committee -- 175 FCT
			// G Suite Service

			// Exchange Committee total: 175 FCT
			HardGrant{221126, 175e8, validateAddress("FA2kd7iAuCrR2GTV39UMaBTphzvQZYVYmvLJYGsjoJRjEGoVNQFd")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Devon Katz-Canonical Ledgers-DBGrow -- 5448 FCT
			// FAT Development 4 - Continuation 2019-12-01 - 2020-02-01

			// Canonical Ledgers total: 2980 FCT
			HardGrant{221126, 2980e8, validateAddress("FA2xccSAfhGm5k4tPaXF9741xkQ52drWjoJodQhpPxDxepdqasMM")},

			// DBGrow total: 2468 FCT
			HardGrant{221126, 2468e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// David Chapman-Motion Factory-Factomize -- 800 FCT
			// PegNet Explainer Video Backpay Grant

			// Motion Factory total: 800 FCT
			HardGrant{221126, 800e8, validateAddress("FA3cRbg1CAzBvXpp8v3AfGuxzjePDbwqGr71MvLNanTX2QKDF1Cz")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Valentin Ganev-Factomatic -- 5500 FCT
			// Python DID library

			// Factomatic total: 5500 FCT
			HardGrant{221126, 5500e8, validateAddress("FA2QuNHNxgJBZyPggxkU8C16YReA4xFFxUvAGpu9azF3TaZg46SF")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Cody B -- 950 FCT
			// Discord Daily Digest Screenshot (Continuation) 2019-11-01 - 2020-01-31

			// Cody B total: 950 FCT
			HardGrant{221126, 950e8, validateAddress("FA2WvyDhuKerk3RyVeSiTUonmcn7RuPRJGFQB6oCfroSys7NW3q2")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Jason Gregoire-Kompendium -- 730 FCT
			// [Back-pay] Further Development of Open API – Client Library & Server Expansion

			// Kompendium total: 730 FCT
			HardGrant{221126, 730e8, validateAddress("FA3KMPNX8AKdY3tjyKMzZ5cAkqUv97d3QqkJeQnVXk6PXSVgArnr")},
			// --------------------------------------------------------

			// ********************************
			// ** BUG Bounty payout 2019-04  **
			// ********************************
			// --------------------------------------------------------
			// Niels Klomp -- Core committee -- 205 FCT
			// This grant is for the Bug Bounty program and pays out 3 submissions, BB-001, BB-002 and BB-003
			// Payout will go to the following addresses, with a managed intermediary address.
			// (Not all exchanges support coinbase transactions)
			//
			// BB-001 -- 17 FCT  -- FA3ShvrkkVrCWGGiGqmAPykhLHhuZZtetUYBTYEss4KPsm5c7G6c
			// BB-002 -- 17 FCT  -- FA29JkXWYzwgLpQcrqSPsouy23AMgbrQcr6PTbAu6Q2tGXNveL4P
			// BB-003 -- 171 FCT -- FA23fEJadueuxf8SsohpnDj3QYmZQ7mpMjWrFb1GHTQhByyDYMxe
			//
			// Full details for core committee:
			// https://docs.google.com/spreadsheets/d/15fo9PCNt4meJGSd-V9iZENAiPTUx7dLpd990tH1dADY/edit#gid=0

			// Bug bounty total: 205 FCT, remaining funds: 795 FCT
			HardGrant{221126, 205e8, validateAddress("FA22J4Age2aLKRw1cKQckTPzK6Wpb6GCKKdYv6dp7SVsxzGyNqy4")},
			// --------------------------------------------------------

			// ********************************
			// **** Grant Round 2020-01 ****
			// ********************************

			// --------------------------------------------------------
			// Samuel Vanderwaal-Brian Deery-Nolan Bauer-Nic R-The 42nd Factoid AS-Centis BV -- 9000 FCT
			// Guide Compensation 2019-12-07 - 2020-03-06

			// Brian Deery total: 1800 FCT
			HardGrant{233201, 1800e8, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// Nolan Bauer total: 1800 FCT
			HardGrant{233201, 1800e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},

			// Nic R total: 1800 FCT
			HardGrant{233201, 1800e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},

			// The 42nd Factoid AS total: 1800 FCT
			HardGrant{233201, 1800e8, validateAddress("FA3AEL2H9XZy3n199USs2poCEJBkK1Egy6JXhLehfLJjUYMKh1zS")},

			// Centis BV total: 1800 FCT
			HardGrant{233201, 1800e8, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// PaulSnow-Factom Inc. -- 900 FCT
			// Oracle Master 2020-03-09 - 2020-06-09

			// Factom Inc. total: 900 FCT
			HardGrant{233201, 900e8, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// PaulSnow-Factom Inc. -- 660 FCT
			// Anchor Master 2020-03-09 - 2020-06-09

			// Factom Inc. total: 660 FCT
			HardGrant{233201, 660e8, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// David Chapman-Factomize -- 19524 FCT
			// Core and General Development 2020-03-01 - 2020-05-31

			// Factomize total: 19524 FCT
			HardGrant{233201, 19524e8, validateAddress("FA3nsSjUy5uSkqMEug8t3VcehZn5w2ciSMpgqFEEsMRwMrHoa9k3")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Julian Fletcher-Taylor-Exchange Committee -- 30000 FCT
			// Exchange Working Group - 001 2020-03-01 - 2020-06-01

			// Exchange Committee total: 30000 FCT
			HardGrant{233201, 30000e8, validateAddress("FA2FrnimQdtuqkB7MFgemGy1RBYNnqZMNEmQ5TAsb5TrkbSNebdm")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// PaulSnow-Nolan Bauer-factomatic-David Kuiper-Factom Inc. -- 34950 FCT
			// Protocol Development 2020-03-09 - 2020-06-09

			// Nolan Bauer total: 300 FCT
			HardGrant{233201, 300e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},

			// factomatic total: 300 FCT
			HardGrant{233201, 300e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},

			// David Kuiper total: 300 FCT
			HardGrant{233201, 300e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},

			// Factom Inc. total: 34050 FCT
			HardGrant{233201, 34050e8, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// CryptoLogic-The Factoid Authority-Crypto Logic-Bedrock Solutions-De Facto -- 2700 FCT
			// Factom Open Node Continuity 2020-03-01 - 2020-05-31

			// The Factoid Authority total: 375 FCT
			HardGrant{233201, 375e8, validateAddress("FA2LV4s7LKA9BTgWaJNvcr9Yq8rpiH2XD3vEPY3nwSiNSrnRgkpK")},

			// Crypto Logic total: 375 FCT
			HardGrant{233201, 375e8, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},

			// Bedrock Solutions total: 975 FCT
			HardGrant{233201, 975e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},

			// De Facto total: 975 FCT
			HardGrant{233201, 975e8, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// David Chapman-Factomize -- 1064 FCT
			// Factomize Forum Software Upgrade 2020-03-01 - 2020-03-16

			// Factomize total: 1064 FCT
			HardGrant{233201, 1064e8, validateAddress("FA3nsSjUy5uSkqMEug8t3VcehZn5w2ciSMpgqFEEsMRwMrHoa9k3")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Valentin Ganev-Factomatic -- 8500 FCT
			// Kambani Extension Backpay 2020-02-03 - 2020-02-03

			// Factomatic total: 8500 FCT
			HardGrant{233201, 8500e8, validateAddress("FA2QuNHNxgJBZyPggxkU8C16YReA4xFFxUvAGpu9azF3TaZg46SF")},
			// --------------------------------------------------------

			// --------------------------------------------------------
			// Matt York -- 900 FCT
			// IPFS-bridge 2020-02-01 - 2020-05-01

			// Matt York total: 900 FCT
			HardGrant{233201, 900e8, validateAddress("FA3PTnet6yGwnCkprszYYfHtwqymqQQWKfFjoRkavxSda15VNtwY")},
			// --------------------------------------------------------

			// ********************************
			// ** BUG Bounty payout 2020-01  **
			// ********************************
			// --------------------------------------------------------
			// Niels Klomp -- Core committee -- 37 FCT
			// This grant is for the Bug Bounty program and pays out 1 submission, BB-004.
			//
			// BB-004 -- 37 FCT  -- FA34H8r61Gt5wTNwgtaUHPvNpma4Kof39uJAYxJ5q9cwqKF2y9Kd
			//
			// Full details for core committee:
			// https://docs.google.com/spreadsheets/d/15fo9PCNt4meJGSd-V9iZENAiPTUx7dLpd990tH1dADY/edit#gid=0

			// Bug bounty total: 37 FCT, remaining funds: 758 FCT
			HardGrant{233201, 37e8, validateAddress("FA34H8r61Gt5wTNwgtaUHPvNpma4Kof39uJAYxJ5q9cwqKF2y9Kd")},
			// --------------------------------------------------------

		}

	default:
		hardcodegrants = []HardGrant{}
	}

	return hardcodegrants
}

func validateAddress(a string) interfaces.IAddress {
	if !primitives.ValidateFUserStr(a) {
		panic(fmt.Sprintf("Bad addr(%s) in grant table", a))
	}
	return factoid.NewAddress(primitives.ConvertUserStrToAddress(a))
}

func CheckGrants() {

	hardcodegrants := GetHardCodedGrants()
	// this used to be in an init block but it turns out COINBASE_PAYOUT_FREQUENCY isn't so
	// constants (changed based on network type) so it had to move here to be valid.
	for i, g := range hardcodegrants { // check every hardcoded grant
		if g.DBh%constants.COINBASE_PAYOUT_FREQUENCY != 1 {
			panic(fmt.Sprintf("Bad grant[%d] payout height for %v", i, g))
		}
	}

}

//return a (possibly empty) of coinbase payouts to be scheduled at this height
func GetGrantPayoutsFor(currentDBHeight uint32) []interfaces.ITransAddress {

	outputs := make([]interfaces.ITransAddress, 0)
	// this is only but temporary, once the hard coded grants are payed this code will go away
	// I can't modify the grant list because in simulation it is shared across nodes so for now I just
	// scan the whole list once every 25 blocks
	// I opted for one list knowing it will have to be different for testnet vs mainnet because making it
	// network sensitive just add complexity to the code.
	// there is no need for activation height because the grants have inherent activation heights per grant
	for _, g := range GetHardCodedGrants() { // check every hardcoded grant
		if g.DBh == currentDBHeight { // if it's ready {...
			o := factoid.NewOutAddress(g.Address, g.Amount) // Create a payout
			outputs = append(outputs, o)                    // and add it to the list
		}
	}
	return outputs
}
