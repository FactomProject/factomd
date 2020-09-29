package p2p

// Info holds the data that can be queried from the Network
type Info struct {
	Peers     int     // number of peers connected
	Receiving float64 // download rate in Messages/s
	Sending   float64 // upload rate in Messages/s
	Download  float64 // download rate in Bytes/s
	Upload    float64 // upload rate in Bytes/s
	Dropped   uint64  // number of parcels dropped due to low speed
}
