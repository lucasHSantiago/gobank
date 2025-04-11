package main

import (
	"database/sql"
	"flag"
	"log"
	"net"

	_ "github.com/lib/pq"
	"github.com/lucasHSantiago/gobank/internal/api"
	db "github.com/lucasHSantiago/gobank/internal/db/sqlc"
	"github.com/lucasHSantiago/gobank/internal/db/util"
	"github.com/lucasHSantiago/gobank/internal/gapi"
	"github.com/lucasHSantiago/gobank/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	var grpc bool
	flag.BoolVar(&grpc, "grpc", false, "Run server as a gRPC")

	flag.Parse()

	println(grpc)

	store := db.NewStore(conn)

	if grpc {
		runGrpcServer(config, store)
	} else {
		runGinServer(config, store)
	}
}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()
	gen.RegisterGoBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GrpcServerAddress)
	if err != nil {
		log.Fatal("cannot create listener")
	}

	log.Printf("start gRPC server at %s\n", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server")
	}
}

// @TODO: change to use flags to run gin or grpc
func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.HttpServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
