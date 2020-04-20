package services

import (
	"context"
	"github.com/autom8ter/userdb/db"
	api "github.com/autom8ter/userdb/gen/go/userdb"
)

func (p *UserDB) GetEmails(ctx context.Context, r *api.GetEmailsRequest) (*api.GetEmailsResponse, error) {
	return &api.GetEmailsResponse{
		Emails: db.GetEmails(p.db),
	}, nil
}

func (p *UserDB) GetRegexEmails(ctx context.Context, r *api.GetRegexEmailsRequest) (*api.GetRegexEmailsResponse, error) {
	keys, err := db.GetRegexEmails(p.db, r.Regex)
	if err != nil {
		return nil, err
	}
	return &api.GetRegexEmailsResponse{
		Emails: keys,
	}, nil
}
