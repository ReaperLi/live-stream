package main

import (
	"github.com/reaper/live-stream/api"
	"log"
)

func main() {
	server := api.NewServer()

	err := server.Start(":9000")
	if err != nil {
		log.Fatal("cannot start http server")
	}
}
