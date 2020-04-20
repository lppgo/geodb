package services

import (
	"context"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/autom8ter/userdb/stream"
	"github.com/dgraph-io/badger/v2"
	"golang.org/x/oauth2"
)

type UserDB struct {
	hub    *stream.Hub
	db     *badger.DB
	config *oauth2.Config
}

func NewUserDB(db *badger.DB, hub *stream.Hub, config *oauth2.Config) *UserDB {
	return &UserDB{
		hub:    hub,
		db:     db,
		config: config,
	}
}

func (p *UserDB) Ping(ctx context.Context, req *api.PingRequest) (*api.PingResponse, error) {
	return &api.PingResponse{
		Ok: true,
	}, nil
}
