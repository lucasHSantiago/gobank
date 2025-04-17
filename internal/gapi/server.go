package gapi

import (
	"fmt"

	db "github.com/lucasHSantiago/gobank/internal/db/sqlc"
	"github.com/lucasHSantiago/gobank/internal/db/util"
	"github.com/lucasHSantiago/gobank/internal/token"
	"github.com/lucasHSantiago/gobank/internal/worker"
	"github.com/lucasHSantiago/gobank/proto/gen"
)

type Server struct {
	gen.UnimplementedGoBankServer
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSecretKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil
}
