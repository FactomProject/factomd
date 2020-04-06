package p2p

import (
	"bytes"
	"testing"
)

func TestMetricsReadWriter_ReadWrite(t *testing.T) {
	// testing of this isn't very rigorous since this basically just a pass-through
	// and metricsRW doesn't touch this

	buf := bytes.NewBuffer(nil)
	mrw := NewMetricsReadWriter(buf)

	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		sc      *MetricsReadWriter
		args    args
		want    int
		wantErr bool
	}{
		{"one byte", mrw, args{[]byte{0x1}}, 1, false},
		{"phrase", mrw, args{[]byte("foo")}, 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sc.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricsReadWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MetricsReadWriter.Write() = %v, want %v", got, tt.want)
			}

			written := make([]byte, len(tt.args.p))
			if n, err := tt.sc.Read(written); err != nil {
				t.Errorf("error reading from buf: %v", err)
			} else if n != len(tt.args.p) {
				t.Errorf("read less than expected, read = %d, wanted = %d", n, len(tt.args.p))
			}

			if bytes.Compare(written, tt.args.p) != 0 {
				t.Errorf("Contents of buffer don't match up with args: %x vs %x", written, tt.args.p)
			}
		})
	}
}

func TestMetricsReadWriter_Collect(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	mrw := NewMetricsReadWriter(buf)

	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		sc      *MetricsReadWriter
		args    args
		want    int
		wantErr bool
	}{
		{"one byte", mrw, args{[]byte{0x1}}, 1, false},
		{"phrase", mrw, args{[]byte("foo")}, 3, false},
		{"zero write", mrw, args{[]byte("")}, 0, false},
		{"longer", mrw, args{[]byte("abcdefghijklmnopqrstuvwxyz")}, 26, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sc.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricsReadWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MetricsReadWriter.Write() = %v, want %v", got, tt.want)
			}

			mw, mr, bw, br := mrw.Collect()
			if br > 0 || mr > 0 {
				t.Errorf("Recorded reads even though no reads. bytes = %d, messages = %d", br, mr)
			}

			if mw != 1 {
				t.Errorf("Writes not as expected. got = %d, want = %d", mw, 1)
			}
			if bw != uint64(len(tt.args.p)) {
				t.Errorf("Byte Writes not as expected. got = %d, want = %d", bw, len(tt.args.p))
			}

			written := make([]byte, len(tt.args.p))
			tt.sc.Read(written)

			mw, mr, bw, br = mrw.Collect()
			if bw > 0 || mw > 0 {
				t.Errorf("Recorded writes even though nothing written. bytes = %d, messages = %d", bw, mw)
			}

			if mr != 1 {
				t.Errorf("Writes not as expected. got = %d, want = %d", mr, 1)
			}
			if br != uint64(len(tt.args.p)) {
				t.Errorf("Byte reads not as expected. got = %d, want = %d", br, len(tt.args.p))
			}

		})
	}

	mrw.Write([]byte("foo"))
	mrw.Write([]byte("bar"))
	mrw.Write([]byte("meow meow meow"))

	five := make([]byte, 5)
	mrw.Read(five)
	mrw.Read(five)

	mw, mr, bw, br := mrw.Collect()
	if mw != 3 {
		t.Errorf("Expected 3 writes, got %d", mw)
	}
	if bw != 20 {
		t.Errorf("Expected 20 bytes written, got %d", bw)
	}

	if mr != 2 {
		t.Errorf("Expected 2 reads, got %d", mr)
	}
	if br != 10 {
		t.Errorf("Expected 10 bytes read, got %d", br)
	}

}
