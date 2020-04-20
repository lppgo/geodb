package services

import (
	"context"
	"github.com/autom8ter/userdb/db"
	api "github.com/autom8ter/userdb/gen/go/userdb"
)

func (p *UserDB) SetUserAccount(ctx context.Context, r *api.SetUserAccountRequest) (*api.SetUserAccountResponse, error) {
	usr, err := db.SetUserAccount(p.db, r.UserEmail, r.AccountName)
	if err != nil {
		return nil, err
	}
	return &api.SetUserAccountResponse{
		User: usr,
	}, nil
}
