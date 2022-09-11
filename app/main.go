package main

import (
	"database/sql"
	"github.com/reaper/live-stream/api"
	db "github.com/reaper/live-stream/db/sqlc"
	"github.com/reaper/live-stream/util"
	"log"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config files")
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)

	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server instance")
	}

	err = server.Start(":9000")
	if err != nil {
		log.Fatal("cannot start http server")
	}
}
