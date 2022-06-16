package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/FactomProject/factomd/Utilities/snapshot/internal/cmd"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	_ = cmd.RootCmd().Execute()
}
