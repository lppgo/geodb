package auth

import (
	"context"
	"fmt"
	"github.com/autom8ter/userdb/config"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/dgrijalva/jwt-go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type GoogleUser struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
}

func BasicAuthFunc() grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		if config.Config.IsSet("USERDB_PASSWORD") {
			basicAuth, err := grpc_auth.AuthFromMD(ctx, "basic")
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "failed to find authentication header with basic scheme\n%v", err)
			}
			if basicAuth != config.Config.GetString("USERDB_PASSWORD") {
				return nil, status.Error(codes.Unauthenticated, "invalid password")
			}
		}
		return ctx, nil
	}
}

func NewUserJWT(userEmail string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(config.Config.GetDuration("USERDB_JWT_EXPIRATION")).Unix(),
		Id:        userEmail,
		IssuedAt:  time.Now().Unix(),
		Issuer:    "userdb",
	})
	return token.SignedString([]byte(config.Config.GetString("USERDB_JWT_SECRET")))
}

func UserFromJWT(tokenString string, fn func(email string) (*api.UserDetail, error)) (*api.UserDetail, error) {
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if !t.Valid {
			return nil, fmt.Errorf("invalid token")
		}
		return config.Config.GetString("USERDB_JWT_SECRET"), nil
	})
	if err != nil {
		return nil, err
	}
	email := claims.Id
	return fn(email)
}
