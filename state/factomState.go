package state

import (
	"github.com/FactomProject/factomd/database"
)

type FactoidState struct {
	DB database.DBOverlay
}
