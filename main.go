package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/lucasHSantiago/gobank/api"
	db "github.com/lucasHSantiago/gobank/db/sqlc"
)

const (
	dbDrive       = "postgres"
	dbSource      = "postgresql://postgres:admin@localhost:5432/gobank?sslmode=disable"
	serverAddress = "localhost:8080"
)

func main() {

	conn, err := sql.Open(dbDrive, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
