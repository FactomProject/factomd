package eventservices

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)
import "github.com/FactomProject/factomd/events/eventmessages/generated/eventmessages"

func mapAdminBlock(block interfaces.IAdminBlock) *eventmessages.AdminBlock {
	result := &eventmessages.AdminBlock{
		Header:  mapAdminBlockHeader(block.GetHeader()),
		Entries: mapAdminBlockEntries(block.GetABEntries()),
	}
	return result
}

func mapAdminBlockHeader(header interfaces.IABlockHeader) *eventmessages.AdminBlockHeader {
	result := &eventmessages.AdminBlockHeader{
		PreviousBackRefHash: &eventmessages.Hash{
			HashValue: header.GetPrevBackRefHash().Bytes(),
		},
		BlockHeight:         header.GetDBHeight(),
		HeaderExpansionSize: header.GetHeaderExpansionSize(),
		HeaderExpansionArea: header.GetHeaderExpansionArea(),
		MessageCount:        header.GetMessageCount(),
		BodySize:            header.GetBodySize(),
	}
	return result
}

func mapAdminBlockEntries(entries []interfaces.IABEntry) []*eventmessages.AdminBlockEntry {
	result := make([]*eventmessages.AdminBlockEntry, len(entries))
	for i, entry := range entries {
		result[i] = &eventmessages.AdminBlockEntry{}
		switch entry.(type) {
		case *adminBlock.AddAuditServer:
			result[i].Value = mapAddAuditServer(entry)
		case *adminBlock.AddEfficiency:
			result[i].Value = mapAddEfficiency(entry)
		case *adminBlock.AddFactoidAddress:
			result[i].Value = mapAddFactoidAddress(entry)
		case *adminBlock.AddFederatedServer:
			result[i].Value = mapAddFederatedServer(entry)
		case *adminBlock.AddFederatedServerBitcoinAnchorKey:
			result[i].Value = mapAddFederatedServerBitcoinAnchorKey(entry)
		case *adminBlock.AddFederatedServerSigningKey:
			result[i].Value = mapAddFederatedServerSigningKey(entry)
		case *adminBlock.AddReplaceMatryoshkaHash:
			result[i].Value = mapAddReplaceMatryoshkaHash(entry)
		case *adminBlock.CancelCoinbaseDescriptor:
			result[i].Value = mapCancelCoinbaseDescriptor(entry)
		case *adminBlock.CoinbaseDescriptor:
			result[i].Value = mapCoinbaseDescriptor(entry)
		case *adminBlock.DBSignatureEntry:
			result[i].Value = mapDBSignatureEntry(entry)
		case *adminBlock.EndOfMinuteEntry:
			result[i].Value = mapEndOfMinuteEntry(entry)
		case *adminBlock.ForwardCompatibleEntry:
			result[i].Value = mapForwardCompatibleEntry(entry)
		case *adminBlock.IncreaseServerCount:
			result[i].Value = mapIncreaseServerCount(entry)
		case *adminBlock.RemoveFederatedServer:
			result[i].Value = mapRemoveFederatedServer(entry)
		case *adminBlock.RevealMatryoshkaHash:
			result[i].Value = mapRevealMatryoshkaHash(entry)
		case *adminBlock.ServerFault:
			result[i].Value = mapServerFault(entry)
		}
	}
	return result
}

func mapAddAuditServer(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddAuditServer {
	addAuditServer, _ := entry.(*adminBlock.AddAuditServer)
	return &eventmessages.AdminBlockEntry_AddAuditServer{
		AddAuditServer: &eventmessages.AddAuditServer{
			IdentityChainID: &eventmessages.Hash{
				HashValue: addAuditServer.IdentityChainID.Bytes(),
			},
			BlockHeight: addAuditServer.DBHeight,
		},
	}
}

func mapAddEfficiency(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddEfficiency {
	addEfficiency, _ := entry.(*adminBlock.AddEfficiency)
	return &eventmessages.AdminBlockEntry_AddEfficiency{
		AddEfficiency: &eventmessages.AddEfficiency{
			IdentityChainID: &eventmessages.Hash{
				HashValue: addEfficiency.IdentityChainID.Bytes(),
			},
			Efficiency: uint32(addEfficiency.Efficiency),
		},
	}
}

func mapAddFactoidAddress(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddFactoidAddress {
	addFactoidAddress, _ := entry.(*adminBlock.AddFactoidAddress)
	return &eventmessages.AdminBlockEntry_AddFactoidAddress{
		AddFactoidAddress: &eventmessages.AddFactoidAddress{
			IdentityChainID: &eventmessages.Hash{
				HashValue: addFactoidAddress.IdentityChainID.Bytes(),
			},
			Address: &eventmessages.Hash{
				HashValue: addFactoidAddress.FactoidAddress.Bytes(),
			},
		},
	}
}

func mapAddFederatedServer(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddFederatedServer {
	addFederatedServer, _ := entry.(*adminBlock.AddFederatedServer)
	return &eventmessages.AdminBlockEntry_AddFederatedServer{
		AddFederatedServer: &eventmessages.AddFederatedServer{
			IdentityChainID: &eventmessages.Hash{
				HashValue: addFederatedServer.IdentityChainID.Bytes(),
			},
		},
	}
}

func mapAddFederatedServerBitcoinAnchorKey(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddFederatedServerBitcoinAnchorKey {
	addBitcoinAnchorKey, _ := entry.(*adminBlock.AddFederatedServerBitcoinAnchorKey)
	return &eventmessages.AdminBlockEntry_AddFederatedServerBitcoinAnchorKey{
		AddFederatedServerBitcoinAnchorKey: &eventmessages.AddFederatedServerBitcoinAnchorKey{
			IdentityChainID: &eventmessages.Hash{
				HashValue: addBitcoinAnchorKey.IdentityChainID.Bytes(),
			},
			KeyPriority:    uint32(addBitcoinAnchorKey.KeyPriority),
			KeyType:        uint32(addBitcoinAnchorKey.KeyType),
			EcdsaPublicKey: addBitcoinAnchorKey.ECDSAPublicKey[:],
		},
	}
}

func mapAddFederatedServerSigningKey(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddFederatedServerSigningKey {
	addSigningKey, _ := entry.(*adminBlock.AddFederatedServerSigningKey)
	return &eventmessages.AdminBlockEntry_AddFederatedServerSigningKey{
		AddFederatedServerSigningKey: &eventmessages.AddFederatedServerSigningKey{
			IdentityChainID: &eventmessages.Hash{
				HashValue: addSigningKey.IdentityChainID.Bytes(),
			},
			KeyPriority: uint32(addSigningKey.KeyPriority),
			PublicKey:   addSigningKey.PublicKey[:],
			BlockHeight: addSigningKey.DBHeight,
		},
	}
}

func mapAddReplaceMatryoshkaHash(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddReplaceMatryoshkaHash {
	addReplaceMatryoshkaHash, _ := entry.(*adminBlock.AddReplaceMatryoshkaHash)
	return &eventmessages.AdminBlockEntry_AddReplaceMatryoshkaHash{
		AddReplaceMatryoshkaHash: &eventmessages.AddReplaceMatryoshkaHash{
			IdentityChainID: &eventmessages.Hash{
				HashValue: addReplaceMatryoshkaHash.IdentityChainID.Bytes(),
			},
			MatryoshkaHash: &eventmessages.Hash{
				HashValue: addReplaceMatryoshkaHash.MHash.Bytes(),
			},
		},
	}
}

func mapCancelCoinbaseDescriptor(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_CancelCoinbaseDescriptor {
	cancelCoinbaseDescriptor, _ := entry.(*adminBlock.CancelCoinbaseDescriptor)
	return &eventmessages.AdminBlockEntry_CancelCoinbaseDescriptor{
		CancelCoinbaseDescriptor: &eventmessages.CancelCoinbaseDescriptor{
			DescriptorHeight: cancelCoinbaseDescriptor.DescriptorHeight,
			DescriptorIndex:  cancelCoinbaseDescriptor.DescriptorIndex,
		},
	}
}

func mapCoinbaseDescriptor(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_CoinbaseDescriptor {
	coinbaseDescriptor, _ := entry.(*adminBlock.CoinbaseDescriptor)
	return &eventmessages.AdminBlockEntry_CoinbaseDescriptor{
		CoinbaseDescriptor: &eventmessages.CoinbaseDescriptor{
			FactoidOutputs: mapTransactionAddresses(coinbaseDescriptor.Outputs),
		},
	}
}

func mapDBSignatureEntry(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_DirectoryBlockSignatureEntry {
	dBSignatureEntry, _ := entry.(*adminBlock.DBSignatureEntry)
	return &eventmessages.AdminBlockEntry_DirectoryBlockSignatureEntry{
		DirectoryBlockSignatureEntry: &eventmessages.DirectoryBlockSignatureEntry{
			IdentityAdminChainID: &eventmessages.Hash{
				HashValue: dBSignatureEntry.IdentityAdminChainID.Bytes(),
			},
			PreviousDirectoryBlockSignature: mapSignature(&dBSignatureEntry.PrevDBSig),
		},
	}
}

func mapEndOfMinuteEntry(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_EndOfMinuteEntry {
	endOfMinuteEntry, _ := entry.(*adminBlock.EndOfMinuteEntry)
	return &eventmessages.AdminBlockEntry_EndOfMinuteEntry{
		EndOfMinuteEntry: &eventmessages.EndOfMinuteEntry{
			MinuteNumber: uint32(endOfMinuteEntry.MinuteNumber),
		},
	}
}

func mapForwardCompatibleEntry(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_ForwardCompatibleEntry {
	forwardCompatibleEntry, _ := entry.(*adminBlock.ForwardCompatibleEntry)
	return &eventmessages.AdminBlockEntry_ForwardCompatibleEntry{
		ForwardCompatibleEntry: &eventmessages.ForwardCompatibleEntry{
			Size_: forwardCompatibleEntry.Size,
			Data:  forwardCompatibleEntry.Data,
		},
	}
}

func mapIncreaseServerCount(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_IncreaseServerCount {
	increaseServerCount, _ := entry.(*adminBlock.IncreaseServerCount)
	return &eventmessages.AdminBlockEntry_IncreaseServerCount{
		IncreaseServerCount: &eventmessages.IncreaseServerCount{
			Amount: uint32(increaseServerCount.Amount),
		},
	}
}

func mapRemoveFederatedServer(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_RemoveFederatedServer {
	removeFederatedServer, _ := entry.(*adminBlock.RemoveFederatedServer)
	return &eventmessages.AdminBlockEntry_RemoveFederatedServer{
		RemoveFederatedServer: &eventmessages.RemoveFederatedServer{
			IdentityChainID: &eventmessages.Hash{
				HashValue: removeFederatedServer.IdentityChainID.Bytes(),
			},
			BlockHeight: removeFederatedServer.DBHeight,
		},
	}
}

func mapRevealMatryoshkaHash(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_RevealMatryoshkaHash {
	revealMatryoshkaHash, _ := entry.(*adminBlock.RevealMatryoshkaHash)
	return &eventmessages.AdminBlockEntry_RevealMatryoshkaHash{
		RevealMatryoshkaHash: &eventmessages.RevealMatryoshkaHash{
			IdentityChainID: &eventmessages.Hash{
				HashValue: revealMatryoshkaHash.IdentityChainID.Bytes(),
			},
			MatryoshkaHash: &eventmessages.Hash{
				HashValue: revealMatryoshkaHash.MHash.Bytes(),
			},
		},
	}
}

func mapServerFault(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_ServerFault {
	serverFault, _ := entry.(*adminBlock.ServerFault)
	return &eventmessages.AdminBlockEntry_ServerFault{
		ServerFault: &eventmessages.ServerFault{
			Timestamp: convertTimeToTimestamp(serverFault.Timestamp.GetTime()),
			ServerID: &eventmessages.Hash{
				HashValue: serverFault.ServerID.Bytes(),
			},
			AuditServerID: &eventmessages.Hash{
				HashValue: serverFault.AuditServerID.Bytes(),
			},
			VmIndex:            uint32(serverFault.VMIndex),
			BlockHeight:        serverFault.DBHeight,
			MessageEntryHeight: serverFault.Height,
			SignatureList:      mapSignatureList(serverFault.SignatureList),
		},
	}
}

func mapSignatureList(signatureList adminBlock.SigList) []*eventmessages.Signature {
	result := make([]*eventmessages.Signature, signatureList.Length)
	for i, signature := range signatureList.List {
		result[i] = mapSignature(signature)
	}
	return result
}

func mapSignature(signature interfaces.IFullSignature) *eventmessages.Signature {
	return &eventmessages.Signature{
		PublicKey: signature.GetKey()[:],
		Signature: signature.GetSignature()[:],
	}
}
