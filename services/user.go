package services

import (
	"context"
	"github.com/autom8ter/userdb/db"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (p *UserDB) Set(ctx context.Context, r *api.SetRequest) (*api.SetResponse, error) {
	if err := r.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	objects, err := db.Set(p.db, p.hub, r.User)
	if err != nil {
		return nil, err
	}
	return &api.SetResponse{
		User: objects,
	}, nil
}

func (p *UserDB) GetRegex(ctx context.Context, r *api.GetRegexRequest) (*api.GetRegexResponse, error) {
	objects, err := db.GetRegex(p.db, r.Regex)
	if err != nil {
		return nil, err
	}
	return &api.GetRegexResponse{
		Users: objects,
	}, nil
}

func (p *UserDB) Get(ctx context.Context, r *api.GetRequest) (*api.GetResponse, error) {
	objects, err := db.Get(p.db, r.Emails)
	if err != nil {
		return nil, err
	}
	return &api.GetResponse{
		Users: objects,
	}, nil
}

func (p *UserDB) Delete(ctx context.Context, r *api.DeleteRequest) (*api.DeleteResponse, error) {
	if err := db.Delete(p.db, r.Emails); err != nil {
		return nil, err
	}
	return &api.DeleteResponse{}, nil
}
