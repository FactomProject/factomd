package identityEntries

//https://github.com/FactomProject/FactomDocs/blob/master/Identity.md

//TODO:
//- Add conversion to human-readable private / public key

// Constants for the prefix part of the identity secret and public keys
const (
	IdentityPrivateKeyPrefix1 = "4db6c9" // prefix of "sk1"
	IdentityPrivateKeyPrefix2 = "4db6e7" // prifix of "sk2"
	IdentityPrivateKeyPrefix3 = "4db705" // prefix of "sk3"
	IdentityPrivateKeyPrefix4 = "4db723" // prefix of "sk4"

	IdentityPublicKeyPrefix1 = "3fbeba" // prefix of "id1"
	IdentityPublicKeyPrefix2 = "3fbed8" // prefix of "id2"
	IdentityPublicKeyPrefix3 = "3fbef6" // prefix of "id3"
	IdentityPublicKeyPrefix4 = "3fbf14" // prefix of "id4"
)

// CheckExternalIDsLength checks that the fields of the external ids match the required lengths
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
