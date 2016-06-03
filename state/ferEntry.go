package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
)

var _ = fmt.Print

type FEREntry struct {
	Version string 			`json:"version"`
	ExpirationHeight uint32 	`json:"exiration_height"`
	ResidentHeight uint32		`json:"resident_height"`
	TargetActivationHeight uint32 	`json:"target_activation_height"`
	Priority uint32			`json:"priority"`
	TargetPrice uint64		`json:"target_price"`
}



// Getter Version
func (this *FEREntry) GetVersion() (string) {
	return this.Version
}

// Setter Version
func (this *FEREntry) SetVersion(passedVersion string) (interfaces.IFEREntry) {
	this.Version = passedVersion
	return this
}




// Getter ExpirationHeight
func (this *FEREntry) GetExpirationHeight() (uint32) {
	return this.ExpirationHeight
}

// Setter ExpirationHeight
func (this *FEREntry) SetExpirationHeight(passedExpirationHeight uint32) (interfaces.IFEREntry) {
	this.ExpirationHeight = passedExpirationHeight
	return this
}





// Getter ResidentHeight
func (this *FEREntry) GetResidentHeight() (uint32) {
	return this.ResidentHeight
}

// Setter ResidentHeight
func (this *FEREntry) SetResidentHeight(passedResidentHeight uint32) (interfaces.IFEREntry) {
	this.ResidentHeight = passedResidentHeight
	return this
}





// Getter TargetActivationHeight
func (this *FEREntry) GetTargetActivationHeight() (uint32) {
	return this.TargetActivationHeight
}

// Setter TargetActivationHeight
func (this *FEREntry) SetTargetActivationHeight(passedTargetActivationHeight uint32) (interfaces.IFEREntry) {
	this.TargetActivationHeight = passedTargetActivationHeight
	return this
}





// Getter Priority
func (this *FEREntry) GetPriority() (uint32) {
	return this.Priority
}

// Setter Priority
func (this *FEREntry) SetPriority(passedPriority uint32) (interfaces.IFEREntry) {
	this.Priority = passedPriority
	return this
}





// Getter TargetPrice
func (this *FEREntry) GetTargetPrice() (uint64) {
	return this.TargetPrice
}

// Setter TargetPrice
func (this *FEREntry) SetTargetPrice(passedTargetPrice uint64) (interfaces.IFEREntry) {
	this.TargetPrice = passedTargetPrice
	return this
}

