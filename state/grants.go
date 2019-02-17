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
			HardGrant{21, 2, validateAddress("FA3oajkmHMfqkNMMShmqpwDThzMCuVrSsBwiXM2kYFVRz3MzxNAJ")}, // Pay Clay 2
			HardGrant{31, 4, validateAddress("FA3Ga2XcaheS5NgQ3q22gBpLgE6tXmPu1GhjdU2FsdN2QPMzKJET")}, // Pay Bob 4
			HardGrant{21, 3, validateAddress("FA3Ga2XcaheS5NgQ3q22gBpLgE6tXmPu1GhjdU2FsdN2QPMzKJET")}, // Pay Bob 3
			HardGrant{11, 1, validateAddress("FA3GH7VEFKqTdJcmwGgDrcY4Xh9njQ4EWiJxhJeim6BCA7QuB388")}, // Pay Bill 1
		}
	case "CUSTOM":
		hardcodegrants = []HardGrant{}

	case "MAIN":
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
			HardGrant{tbd, 1200e8, validateAddress("FA2hvRaci9Kks9cLNkEUFcxzUJuUFaaAE1eWYLqa2qk1k9pVFVBp")},

			// The 42nd Factoid total: 1800 FCT
			HardGrant{tbd, 1800e8, validateAddress("FA3AEL2H9XZy3n199USs2poCEJBkK1Egy6JXhLehfLJjUYMKh1zS")},

			// Factom, Inc. total: 1800 FCT
			HardGrant{tbd, 1800e8, validateAddress("FA2teRURMYTdYAA97zdh7rZDkxNtR1nhjryo34aaskjYqsqRSwZq")},

			// Canonical Ledgers total: 1800 FCT
			HardGrant{tbd, 1800e8, validateAddress("FA2PEXgRiPd14NzUP47XfVTgEnvjtLSebBZvnM8gM7cJAMuqWs89")},

			// DBGrow total: 1800 FCT
			HardGrant{tbd, 1800e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},
			// --------------------------------------------------------


			// Anchor and Oracle master grants

			// Factom, Inc 007 -- Oracle Master Dec 9 2018 - Mar 9 2019
			// 300 FCT per month for three months = 900 FCT
			HardGrant{tbd, 900e8, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},

			// Factom, Inc 010 -- Oracle Master Mar 9 2019 - June 9 2019
			// 300 FCT per month for three months = 900 FCT
			HardGrant{tbd, 900e8, validateAddress("FA3fpiZ91MCRRFjVGfNXK4pg7vx3BT3aSRyoVqgptZCX7N5BNR8P")},

			// Factom, Inc 008 -- Anchor Master Dec 9 2018 - Mar 9 2019
			// 220 FCT per month for three months = 660 FCT
			HardGrant{tbd, 660e8, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},

			// Factom, Inc 011 -- Anchor Master Mar 9 2019 - June 9 2019
			// 220 FCT per month for three months = 660 FCT
			HardGrant{tbd, 660e8, validateAddress("FA3jySUFtLXb1VdAJJ5NRVNYEtZ4EBSkDB7yn6LuKGQ4P1ntARhx")},
			// --------------------------------------------------------


			// Committee Grants

			// FACTOM-GRANT-FACTOM-MARKETING-COMMITTEE-002
			// Factom Protocol Marketing Committee - Explainer Video Grant
			HardGrant{tbd, 3000e8, validateAddress("FA3QcDsGS2pK6LDuqCuA2i1sRVeH59PhdHPeQ6bneMDj7ZAqbsZg")},

			// FACTOM-GRANT-FACTOM-MARKETING-COMMITTEE-003
			// Factom Protocol Marketing Committee - Hackathon Grant
			HardGrant{tbd, 3600e8, validateAddress("FA2pdi4o4qJd2y7ygHbyvJTk6oXrPLJvH27zdLZrUkmc9gT6Mpca")},
			// --------------------------------------------------------

			// FACTOM-GRANT-EXCHANGE-COMMITTEE-001
			// Exchange Committee Funding (held back on request in 2018-2 round, paid out during 2019-1 round, see above approved grant proposal)
			HardGrant{tbd, 4276e8, validateAddress("FA3YtPXAJehncSQtA8CWgNSWPr5pyeqovGQb99uVdmdeRqKAeg24")},

			// FACTOM-GRANT-EXCHANGE-COMMITTEE-003
			// Exchange Committee Funding (2019-1 round)
			HardGrant{tbd, 5000e8, validateAddress("FA2feHES9FUQwSDYHeGT8UasUKAPNb91fMDTi7qqpXqHedrKyDwv")},
			// --------------------------------------------------------


			// Core Development

			// Factom Inc, Protocol Development Continuation 009 Dec 9 2018 - Mar 9 2019
			HardGrant{tbd, 27440e8, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")},
			// Sponsor 1, Dominic Luxford -- 600 FCT
			HardGrant{tbd, 600e8, validateAddress("FA27Y2fEsaBPeFsN87czeZxLsA9fxi3fcy4f4xHXdF58W7TgbaCB")},
			// Sponsor 2, Nolan Bauer -- 600 FCT
			HardGrant{tbd, 600e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},
			// Sponsor 3, Factomatic -- 600 FCT
			HardGrant{tbd, 600e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},
			// Transition assistent, Nic Robinette -- 200 FCT
			HardGrant{tbd, 200e8, validateAddress("FA2TDwdBLoCtEKrWqf7gSDhXzw8F5GLTK2wFLgg3erC1Ms6jBWuk")},
			// --------------------------------------------------------

			// Factom Inc, Protocol Development 012 Mar 9 2019 - Jun 9 2019
			HardGrant{tbd, 35459e8, validateAddress("FA3LwCDE3ZdFkr9nE1Keb5JcHgwXVWpEHydshT1x2qKFdvZELVQz")},
			// Sponsor 1, Dominic Luxford -- 600 FCT
			HardGrant{tbd, 600e8, validateAddress("FA27Y2fEsaBPeFsN87czeZxLsA9fxi3fcy4f4xHXdF58W7TgbaCB")},
			// Sponsor 2, Nolan Bauer -- 600 FCT
			HardGrant{tbd, 600e8, validateAddress("FA2oecgJW3XWnXzHhQQoULmMeKC97uAgHcPd4kEowTb3csVkbDc9")},
			// Sponsor 3, Factomatic -- 600 FCT
			HardGrant{tbd, 600e8, validateAddress("FA2944TXTDQKdJDp3TLSANjgMjwK2pQnTSkzE3kQcHWKetCCphcH")},
			// --------------------------------------------------------

			// FACTOM-GRANT-FACTOMIZE-002
			// Factomize, Core Code Development Grant
			HardGrant{tbd, 6000e8, validateAddress("FA3XkRCucFVp2ZMnY5uSkxrzKojirkeY6KpwkJyNZPRJ4LsjmFDp")},
			// --------------------------------------------------------

			// FACTOM-GRANT-LAYERTECH-001
			// LayerTech, Core Code Development Grant
			HardGrant{tbd, 5500e8, validateAddress("FA2qGCTMiufU1cStopyx3NbNwG1Sawpo8MM9icvKXouzA6mSsFbA")},
			// --------------------------------------------------------

			// FACTOM-GRANT-BIF-004 BIF
			// BIF, Core Code Development Grant
			HardGrant{tbd, 500e8, validateAddress("FA2YFZrZkywe1TvSrsWCaZ5RyJ1ZXcN5a2x9WqCRobj5GQawpTNt")},
			// --------------------------------------------------------


			// Factom Open API

			// FACTOM-GRANT-DEFACTO-001
			// Factom Open API — Grant #1
			// De Facto
			HardGrant{tbd, 6980e8, validateAddress("FA2rrwFVvkFYwyGFHVBMwRqTpycuZiagrQdcbPWzuoEwJQxjDwi3")},
			// Jay Cheroske (Bedrock Solutions)
			HardGrant{tbd, 400e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},
			// --------------------------------------------------------


			// Factom Open Node

			// FACTOM-GRANT-BEDROCK-CRYPTLOGIC-DEFACTO-TFA-002
			// Factom Open Node (ex. Courtesy Node) Continuity
			// Bedrock Solutions -- 136.25 FCT
			HardGrant{tbd, 13625e6, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},
			// De Facto -- 136.25 FCT
			HardGrant{tbd, 13625e6, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			// CryptoLogic -- 136.25 FCT
			HardGrant{tbd, 13625e6, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},
			// TFA -- 136.25 FCT
			HardGrant{tbd, 13625e6, validateAddress("FA2LV4s7LKA9BTgWaJNvcr9Yq8rpiH2XD3vEPY3nwSiNSrnRgkpK")},
			// --------------------------------------------------------

			// BEDROCK-DEFACTO-001
			// Factom Open Node Enhancement -- 374 FCT total
			// Bedrock Solutions -- 172 FCT
			HardGrant{tbd, 172e8, validateAddress("FA2FqYZPfBeRWq7fWSFEhassT5zpMQZm8jwus3yWbzeN3PZPWybm")},
			// De Facto -- 172 FCT
			HardGrant{tbd, 172e8, validateAddress("FA2YeMbN8Z1SsT7Yqw6Np85kWwtFVg2CyJKMDFnuXTawWuWPtzvX")},
			// CryptoLogic -- 30 FCT
			HardGrant{tbd, 30e8, validateAddress("FA29wMUjN38BVLbJs6dR6gHHdBys2mpo3wy565JCjquUQTGqNZfb")},
			// --------------------------------------------------------


			// Blockchain Expo Global 2019

			// FACTOM-GRANT-PRESTIGE_IT-001
			// Prestige IT - Blockchain Expo Global 2019 (London)
			HardGrant{tbd, 682e8, validateAddress("FA3iRzBGA78gkkJ88PinKi3wwNfBhyoGExgzYx9btJZqo5or1o5A")},
			// --------------------------------------------------------


			// Marketing Videos

			// FACTOM-GRANT-FACTOMIZE-001
			// Factomize, Marketing videos
			HardGrant{tbd, 500e8, validateAddress("FA3XkRCucFVp2ZMnY5uSkxrzKojirkeY6KpwkJyNZPRJ4LsjmFDp")},
			// --------------------------------------------------------


			// FAT protocol

			// FACTOM-GRANT-DBGrow-Luciap-Canonical Ledgers-002
			// FAT Protocol Development Grant II - 12.700 FCT
			// DBGrow -- 5.500 FCT
			HardGrant{tbd, 5500e8, validateAddress("FA3HSuFo9Soa5ZnG82JHqyKiRi4Pw17LxPTo9AsCaFNLCGkXkgsu")},
			// Luciap -- 3.200 FCT
			HardGrant{tbd, 3200e8, validateAddress("FA3DikVW7pzhMkJXuP9xszf9o3aKrMHqEpPkLee2Nb6WewhupyM8")},
			// Canonical Ledgers -- 4.000 FCT
			HardGrant{tbd, 4000e8, validateAddress("FA2nBeBX75R7ECdhZS61DLpP5apaS32zwSYQ7aRkahwAjy5bryFo")},

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
