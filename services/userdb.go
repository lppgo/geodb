package services

import (
	"context"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/autom8ter/userdb/stream"
	"github.com/dgraph-io/badger/v2"
)

type UserDB struct {
	hub *stream.Hub
	db  *badger.DB
}

func NewUserDB(db *badger.DB, hub *stream.Hub) *UserDB {
	return &UserDB{
		hub: hub,
		db:  db,
	}
}

func (p *UserDB) Ping(ctx context.Context, req *api.PingRequest) (*api.PingResponse, error) {
	return &api.PingResponse{
		Ok: true,
	}, nil
}
