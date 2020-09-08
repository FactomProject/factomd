package eventservices

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)
import "github.com/FactomProject/factomd/modules/livefeed/eventmessages/generated/eventmessages"

func MapAdminBlock(block interfaces.IAdminBlock) *eventmessages.AdminBlock {
	var keyMRBytes []byte = nil
	if keyMR, _ := block.GetKeyMR(); keyMR != nil {
		keyMRBytes = keyMR.Bytes()
	}

	result := &eventmessages.AdminBlock{
		Header:        mapAdminBlockHeader(block.GetHeader()),
		Entries:       mapAdminBlockEntries(block.GetABEntries()),
		KeyMerkleRoot: keyMRBytes,
	}
	return result
}

func mapAdminBlockHeader(header interfaces.IABlockHeader) *eventmessages.AdminBlockHeader {
	result := &eventmessages.AdminBlockHeader{
		PreviousBackRefHash: header.GetPrevBackRefHash().Bytes(),
		BlockHeight:         header.GetDBHeight(),
		MessageCount:        header.GetMessageCount(),
	}
	return result
}

func mapAdminBlockEntries(entries []interfaces.IABEntry) []*eventmessages.AdminBlockEntry {
	result := make([]*eventmessages.AdminBlockEntry, len(entries))
	for i, entry := range entries {
		result[i] = &eventmessages.AdminBlockEntry{}
		switch entry.(type) {
		case *adminBlock.AddAuditServer:
			result[i].AdminBlockEntry = mapAddAuditServer(entry)
		case *adminBlock.AddEfficiency:
			result[i].AdminBlockEntry = mapAddEfficiency(entry)
		case *adminBlock.AddFactoidAddress:
			result[i].AdminBlockEntry = mapAddFactoidAddress(entry)
		case *adminBlock.AddFederatedServer:
			result[i].AdminBlockEntry = mapAddFederatedServer(entry)
		case *adminBlock.AddFederatedServerBitcoinAnchorKey:
			result[i].AdminBlockEntry = mapAddFederatedServerBitcoinAnchorKey(entry)
		case *adminBlock.AddFederatedServerSigningKey:
			result[i].AdminBlockEntry = mapAddFederatedServerSigningKey(entry)
		case *adminBlock.AddReplaceMatryoshkaHash:
			result[i].AdminBlockEntry = mapAddReplaceMatryoshkaHash(entry)
		case *adminBlock.CancelCoinbaseDescriptor:
			result[i].AdminBlockEntry = mapCancelCoinbaseDescriptor(entry)
		case *adminBlock.CoinbaseDescriptor:
			result[i].AdminBlockEntry = mapCoinbaseDescriptor(entry)
		case *adminBlock.DBSignatureEntry:
			result[i].AdminBlockEntry = mapDBSignatureEntry(entry)
		case *adminBlock.EndOfMinuteEntry:
			result[i].AdminBlockEntry = mapEndOfMinuteEntry(entry)
		case *adminBlock.ForwardCompatibleEntry:
			result[i].AdminBlockEntry = mapForwardCompatibleEntry(entry)
		case *adminBlock.IncreaseServerCount:
			result[i].AdminBlockEntry = mapIncreaseServerCount(entry)
		case *adminBlock.RemoveFederatedServer:
			result[i].AdminBlockEntry = mapRemoveFederatedServer(entry)
		case *adminBlock.RevealMatryoshkaHash:
			result[i].AdminBlockEntry = mapRevealMatryoshkaHash(entry)
		case *adminBlock.ServerFault:
			result[i].AdminBlockEntry = mapServerFault(entry)
		}
		result[i].AdminIdType = uint32(entry.Type())
	}
	return result
}

func mapAddAuditServer(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddAuditServer {
	addAuditServer, _ := entry.(*adminBlock.AddAuditServer)
	return &eventmessages.AdminBlockEntry_AddAuditServer{
		AddAuditServer: &eventmessages.AddAuditServer{
			IdentityChainID: addAuditServer.IdentityChainID.Bytes(),
			BlockHeight:     addAuditServer.DBHeight,
		},
	}
}

func mapAddEfficiency(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddEfficiency {
	addEfficiency, _ := entry.(*adminBlock.AddEfficiency)
	return &eventmessages.AdminBlockEntry_AddEfficiency{
		AddEfficiency: &eventmessages.AddEfficiency{
			IdentityChainID: addEfficiency.IdentityChainID.Bytes(),
			Efficiency:      uint32(addEfficiency.Efficiency),
		},
	}
}

func mapAddFactoidAddress(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddFactoidAddress {
	addFactoidAddress, _ := entry.(*adminBlock.AddFactoidAddress)
	return &eventmessages.AdminBlockEntry_AddFactoidAddress{
		AddFactoidAddress: &eventmessages.AddFactoidAddress{
			IdentityChainID: addFactoidAddress.IdentityChainID.Bytes(),
			Address:         addFactoidAddress.FactoidAddress.Bytes(),
		},
	}
}

func mapAddFederatedServer(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddFederatedServer {
	addFederatedServer, _ := entry.(*adminBlock.AddFederatedServer)
	return &eventmessages.AdminBlockEntry_AddFederatedServer{
		AddFederatedServer: &eventmessages.AddFederatedServer{
			IdentityChainID: addFederatedServer.IdentityChainID.Bytes(),
		},
	}
}

func mapAddFederatedServerBitcoinAnchorKey(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddFederatedServerBitcoinAnchorKey {
	addBitcoinAnchorKey, _ := entry.(*adminBlock.AddFederatedServerBitcoinAnchorKey)
	return &eventmessages.AdminBlockEntry_AddFederatedServerBitcoinAnchorKey{
		AddFederatedServerBitcoinAnchorKey: &eventmessages.AddFederatedServerBitcoinAnchorKey{
			IdentityChainID: addBitcoinAnchorKey.IdentityChainID.Bytes(),
			KeyPriority:     uint32(addBitcoinAnchorKey.KeyPriority),
			KeyType:         uint32(addBitcoinAnchorKey.KeyType),
			EcdsaPublicKey:  addBitcoinAnchorKey.ECDSAPublicKey[:],
		},
	}
}

func mapAddFederatedServerSigningKey(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddFederatedServerSigningKey {
	addSigningKey, _ := entry.(*adminBlock.AddFederatedServerSigningKey)
	return &eventmessages.AdminBlockEntry_AddFederatedServerSigningKey{
		AddFederatedServerSigningKey: &eventmessages.AddFederatedServerSigningKey{
			IdentityChainID: addSigningKey.IdentityChainID.Bytes(),
			KeyPriority:     uint32(addSigningKey.KeyPriority),
			PublicKey:       addSigningKey.PublicKey[:],
			BlockHeight:     addSigningKey.DBHeight,
		},
	}
}

func mapAddReplaceMatryoshkaHash(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_AddReplaceMatryoshkaHash {
	addReplaceMatryoshkaHash, _ := entry.(*adminBlock.AddReplaceMatryoshkaHash)
	return &eventmessages.AdminBlockEntry_AddReplaceMatryoshkaHash{
		AddReplaceMatryoshkaHash: &eventmessages.AddReplaceMatryoshkaHash{
			IdentityChainID: addReplaceMatryoshkaHash.IdentityChainID.Bytes(),
			MatryoshkaHash:  addReplaceMatryoshkaHash.MHash.Bytes(),
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
			IdentityAdminChainID:            dBSignatureEntry.IdentityAdminChainID.Bytes(),
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
			IdentityChainID: removeFederatedServer.IdentityChainID.Bytes(),
			BlockHeight:     removeFederatedServer.DBHeight,
		},
	}
}

func mapRevealMatryoshkaHash(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_RevealMatryoshkaHash {
	revealMatryoshkaHash, _ := entry.(*adminBlock.RevealMatryoshkaHash)
	return &eventmessages.AdminBlockEntry_RevealMatryoshkaHash{
		RevealMatryoshkaHash: &eventmessages.RevealMatryoshkaHash{
			IdentityChainID: revealMatryoshkaHash.IdentityChainID.Bytes(),
			MatryoshkaHash:  revealMatryoshkaHash.MHash.Bytes(),
		},
	}
}

func mapServerFault(entry interfaces.IABEntry) *eventmessages.AdminBlockEntry_ServerFault {
	serverFault, _ := entry.(*adminBlock.ServerFault)
	return &eventmessages.AdminBlockEntry_ServerFault{
		ServerFault: &eventmessages.ServerFault{
			Timestamp:          ConvertTimeToTimestamp(serverFault.Timestamp.GetTime()),
			ServerID:           serverFault.ServerID.Bytes(),
			AuditServerID:      serverFault.AuditServerID.Bytes(),
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
