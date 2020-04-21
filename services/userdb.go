package services

import (
	"context"
	"encoding/json"
	"github.com/autom8ter/userdb/auth"
	"github.com/autom8ter/userdb/db"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/dgraph-io/badger/v2"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (p *UserDB) Set(ctx context.Context, r *api.SetRequest) (*api.SetResponse, error) {
	if err := r.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	objects, err := db.Set(p.db, r.User)
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
	dbUser, err := db.Login(p.db, usr)
	if err != nil {
		return nil, err
	}
	jwt, err := auth.NewUserJWT(dbUser.Email)
	if err != nil {
		return nil, err
	}
	return &api.LoginResponse{
		User: dbUser,
		Jwt:  jwt,
	}, nil
}

func (p *UserDB) LoginJWT(ctx context.Context, r *api.LoginJWTRequest) (*api.LoginJWTResponse, error) {
	usr, err := auth.UserFromJWT(r.Jwt, func(email string) (detail *api.UserDetail, err error) {
		usr, err := db.Get(p.db, []string{email})
		if err != nil {
			return nil, err
		}
		return usr[email], nil
	})
	if err != nil {
		return nil, err
	}
	return &api.LoginJWTResponse{
		User: usr,
	}, nil
}

func (p *UserDB) SetSource(ctx context.Context, r *api.SetSourceRequest) (*api.SetSourceResponse, error) {
	detail, err := db.SetSource(p.db, r.Email, r.Source)
	if err != nil {
		return nil, err
	}
	return &api.SetSourceResponse{
		User: detail,
	}, nil
}

func (p *UserDB) UpdateCharge(ctx context.Context, r *api.UpdateChargeRequest) (*api.UpdateChargeResponse, error) {
	charge, err := db.UpdateCharge(r.ChargeID, r.Description, r.Amount, r.Metadata)
	if err != nil {
		return nil, err
	}
	return &api.UpdateChargeResponse{
		ChargeId: charge,
	}, nil
}

func (p *UserDB) RefundCharge(ctx context.Context, r *api.RefundChargeRequest) (*api.RefundChargeResponse, error) {
	_, err := db.RefundCharge(r.ChargeID)
	if err != nil {
		return nil, err
	}
	return &api.RefundChargeResponse{
		Refunded: true,
	}, nil
}

func (p *UserDB) IncPlanUsage(ctx context.Context, r *api.IncPlanUsageRequest) (*api.IncPlanUsageResponse, error) {
	if err := db.IncPlanUsage(p.db, r.Increment, r.Email, r.Plan); err != nil {
		return nil, err
	}
	return &api.IncPlanUsageResponse{}, nil
}

func (p *UserDB) Charge(ctx context.Context, r *api.ChargeRequest) (*api.ChargeResponse, error) {
	chargeID, err := db.Charge(p.db, r.Amount, r.Email, r.Description, r.Metadata)
	if err != nil {
		return nil, err
	}
	return &api.ChargeResponse{
		ChargeId: chargeID,
	}, nil
}

func (p *UserDB) AddPlan(ctx context.Context, r *api.SetPlanRequest) (*api.SetPlanResponse, error) {
	detail, err := db.AddPlan(p.db, r.Email, r.Plan)
	if err != nil {
		return nil, err
	}
	return &api.SetPlanResponse{
		User: detail,
	}, nil
}

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
