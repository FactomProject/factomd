package p2p

import "testing"

func TestV9Msg_Valid(t *testing.T) {
	tests := []struct {
		name    string
		msg     V9Msg
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.Valid(); (err != nil) != tt.wantErr {
				t.Errorf("V9Msg.Valid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
