package identityEntries

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md

//TODO:
//- Add conversion to human-readable private / public key

const (
	IdentityPrivateKeyPrefix1 = "4db6c9"
	IdentityPrivateKeyPrefix2 = "4db6e7"
	IdentityPrivateKeyPrefix3 = "4db705"
	IdentityPrivateKeyPrefix4 = "4db723"

	IdentityPublicKeyPrefix1 = "3fbeba"
	IdentityPublicKeyPrefix2 = "3fbed8"
	IdentityPublicKeyPrefix3 = "3fbef6"
	IdentityPublicKeyPrefix4 = "3fbf14"
)

// Checking the external ids if they match the needed lengths
func CheckExternalIDsLength(extIDs [][]byte, lengths []int) bool {
	if len(extIDs) != len(lengths) {
		return false
	}
	for i := range extIDs {
		if lengths[i] != len(extIDs[i]) {
			return false
		}
	}
	return true
}
