package p2p

import "testing"

func TestParcel_Test(t *testing.T) {
	//&Parcel{Type: 0, Address: "", Payload: []byte{}},
	tests := []struct {
		name    string
		p       *Parcel
		wantErr bool
		app     bool
	}{
		{"blank parcel", &Parcel{}, true, false},
		{"heartbeat", &Parcel{Type: TypeHeartbeat, Address: "", Payload: []byte{0}}, false, false},
		{"empty payload", &Parcel{Type: TypeHeartbeat, Address: "", Payload: []byte{}}, true, false},
		{"ping", &Parcel{Type: TypePing, Address: "", Payload: []byte{0}}, false, false},
		{"pong", &Parcel{Type: TypePong, Address: "", Payload: []byte{0}}, false, false},
		{"p request", &Parcel{Type: TypePeerRequest, Address: "", Payload: []byte{0}}, false, false},
		{"p response", &Parcel{Type: TypePeerResponse, Address: "", Payload: []byte{0}}, false, false},
		{"alert", &Parcel{Type: TypeAlert, Address: "", Payload: []byte{0}}, false, false},
		{"message", &Parcel{Type: TypeMessage, Address: "", Payload: []byte{0}}, false, true},
		{"messagepart", &Parcel{Type: TypeMessagePart, Address: "", Payload: []byte{0}}, false, true},
		{"handshake", &Parcel{Type: TypeHandshake, Address: "", Payload: []byte{0}}, false, false},
		{"reject alternative", &Parcel{Type: TypeRejectAlternative, Address: "", Payload: []byte{0}}, false, false},
		{"out of range", &Parcel{Type: ParcelType(len(typeStrings)), Address: "", Payload: []byte{0}}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Valid(); (err != nil) != tt.wantErr {
				t.Errorf("Parcel.Valid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if isapp := tt.p.IsApplicationMessage(); isapp != tt.app {
				t.Errorf("Parcel.IsApplicationMessage() = %v, app = %v", isapp, tt.app)
			}

		})
	}
}
