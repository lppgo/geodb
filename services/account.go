package services

import (
	"context"
	"github.com/autom8ter/userdb/db"
	api "github.com/autom8ter/userdb/gen/go/userdb"
)

func (p *UserDB) NewAccount(ctx context.Context, r *api.NewAccountRequest) (*api.NewAccountResponse, error) {
	acc, err := db.NewAccount(p.db, r.Name, r.AdminEmail, r.Metadata)
	if err != nil {
		return nil, err
	}
	return &api.NewAccountResponse{
		Account: acc,
	}, nil
}

func (p *UserDB) GetAccount(ctx context.Context, r *api.GetAccountRequest) (*api.GetAccountResponse, error) {
	acc, err := db.GetAccount(p.db, r.Names)
	if err != nil {
		return nil, err
	}
	return &api.GetAccountResponse{
		Accounts: acc,
	}, nil
}

func (p *UserDB) GetAccountRegex(ctx context.Context, r *api.GetAccountRegexRequest) (*api.GetAccountRegexResponse, error) {
	acc, err := db.GetAccountRegex(p.db, r.Regex)
	if err != nil {
		return nil, err
	}
	return &api.GetAccountRegexResponse{
		Accounts: acc,
	}, nil
}

func (p *UserDB) DeleteAccount(ctx context.Context, r *api.DeleteAccountRequest) (*api.DeleteAccountResponse, error) {
	err := db.DeleteAccount(p.db, r.Names)
	if err != nil {
		return nil, err
	}
	return &api.DeleteAccountResponse{}, nil
}

func (p *UserDB) AddAccountPlan(ctx context.Context, r *api.SetAccountPlanRequest) (*api.SetAccountPlanResponse, error) {
	acc, err := db.AddAccountPlan(p.db, r.AccountName, r.Plan)
	if err != nil {
		return nil, err
	}
	return &api.SetAccountPlanResponse{
		Account: acc,
	}, nil
}

func (p *UserDB) SetAccountSource(ctx context.Context, r *api.SetAccountSourceRequest) (*api.SetAccountSourceResponse, error) {
	acc, err := db.SetAccountSource(p.db, r.AccountName, r.Source)
	if err != nil {
		return nil, err
	}
	return &api.SetAccountSourceResponse{
		Account: acc,
	}, nil
}

func (p *UserDB) GetAccountNames(ctx context.Context, r *api.GetAccountNamesRequest) (*api.GetAccountNamesResponse, error) {
	acc, err := db.GetAccountNames(p.db)
	if err != nil {
		return nil, err
	}
	return &api.GetAccountNamesResponse{
		Names: acc,
	}, nil
}

func (p *UserDB) GetAccountNamesRegex(ctx context.Context, r *api.GetAccountNamesRegexRequest) (*api.GetAccountNamesRegexResponse, error) {
	acc, err := db.GetAccountNamesRegex(p.db, r.Regex)
	if err != nil {
		return nil, err
	}
	return &api.GetAccountNamesRegexResponse{
		Names: acc,
	}, nil
}

func (p *UserDB) IncAccountPlanUsage(ctx context.Context, r *api.IncAccountPlanUsageRequest) (*api.IncAccountPlanUsageResponse, error) {
	if err := db.IncAccountPlanUsage(p.db, r.Increment, r.AccountName, r.Plan); err != nil {
		return nil, err
	}
	return &api.IncAccountPlanUsageResponse{}, nil
}

func (p *UserDB) ChargeAccount(ctx context.Context, r *api.ChargeAccountRequest) (*api.ChargeAccountResponse, error) {
	chargeID, err := db.ChargeAccount(p.db, r.Amount, r.AccountName, r.Description, r.Metadata)
	if err != nil {
		return nil, err
	}
	return &api.ChargeAccountResponse{
		ChargeId: chargeID,
	}, nil
}
