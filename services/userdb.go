package services

import (
	"context"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/dgraph-io/badger/v2"
	"golang.org/x/oauth2"
)

type UserDB struct {
	db     *badger.DB
	config *oauth2.Config
}

func NewUserDB(db *badger.DB, config *oauth2.Config) *UserDB {
	return &UserDB{
		db:     db,
		config: config,
	}
}

func (p *UserDB) Ping(ctx context.Context, req *api.PingRequest) (*api.PingResponse, error) {
	return &api.PingResponse{
		Ok: true,
	}, nil
}
