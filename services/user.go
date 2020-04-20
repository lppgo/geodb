package services

import (
	"context"
	"encoding/json"
	"github.com/autom8ter/userdb/auth"
	"github.com/autom8ter/userdb/db"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"golang.org/x/oauth2"
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

func (p *UserDB) SetSource(ctx context.Context, r *api.SetSourceRequest) (*api.SetSourceResponse, error) {
	if err := r.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	objects, err := db.SetSource(p.db, p.hub, r.Email, r.Source)
	if err != nil {
		return nil, err
	}
	return &api.SetSourceResponse{
		User: objects,
	}, nil
}

func (p *UserDB) SetPlan(ctx context.Context, r *api.SetPlanRequest) (*api.SetPlanResponse, error) {
	if err := r.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	objects, err := db.SetPlan(p.db, p.hub, r.Email, r.Plan)
	if err != nil {
		return nil, err
	}
	return &api.SetPlanResponse{
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

func (p *UserDB) Login(ctx context.Context, r *api.LoginRequest) (*api.LoginResponse, error) {
	const authURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	token, err := p.config.Exchange(ctx, r.Code, oauth2.AccessTypeOnline)
	if err != nil {
		return nil, err
	}
	client := p.config.Client(ctx, token)
	resp, err := client.Get(authURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	usr := &auth.GoogleUser{}
	if err := json.NewDecoder(resp.Body).Decode(usr); err != nil {
		return nil, err
	}
	dbUser, err := db.Login(p.db, p.hub, usr)
	if err != nil {
		return nil, err
	}
	return &api.LoginResponse{
		User: dbUser,
	}, nil
}
