package main

import (
	"database/sql"
	"flag"
	"github.com/reaper/live-stream/api"
	db "github.com/reaper/live-stream/db/sqlc"
	"github.com/reaper/live-stream/util"
	"log"
)

func main() {
	port := flag.String("port", "9000", "the server port")
	flag.Parse()
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

	err = server.Start(":" + *port)
	if err != nil {
		log.Fatal("cannot start http server")
	}
}
