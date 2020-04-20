package server

import (
	"context"
	"fmt"
	"github.com/autom8ter/userdb/auth"
	"github.com/autom8ter/userdb/config"
	"github.com/dgraph-io/badger/v2"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/piotrkowalczuk/promgrpc/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	stripe2 "github.com/stripe/stripe-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"time"
)

type Server struct {
	server     *grpc.Server
	router     *echo.Echo
	db         *badger.DB
	hTTPClient *http.Client
	logger     *log.Logger
	config     *oauth2.Config
}

func (s *Server) GetGRPCServer() *grpc.Server {
	return s.server
}

func (s *Server) GetDB() *badger.DB {
	return s.db
}

func (s *Server) GetOAuth() *oauth2.Config {
	return s.config
}

func (s *Server) GetHTTPClient() *http.Client {
	return s.hTTPClient
}

func (s *Server) GetLogger() *log.Logger {
	return s.logger
}

func GetDeps() (*badger.DB, *oauth2.Config, error) {
	stripe2.Key = config.Config.GetString("USERDB_STRIPE_KEY")
	db, err := badger.Open(badger.DefaultOptions(config.Config.GetString("USERDB_PATH")))
	if err != nil {
		return nil, nil, err
	}
	cfg := &oauth2.Config{
		ClientID:     config.Config.GetString("USERDB_GOOGLE_CLIENT_ID"),
		ClientSecret: config.Config.GetString("USERDB_GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  config.Config.GetString("USERDB_GOOGLE_REDIRECT"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	}
	return db, cfg, nil
}

func NewServer() (*Server, error) {
	db, cfg, err := GetDeps()
	if err != nil {
		return nil, err
	}
	var promInterceptor = promgrpc.NewInterceptor(promgrpc.InterceptorOpts{})
	if err := prometheus.DefaultRegisterer.Register(promInterceptor); err != nil {
		return nil, err
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_ctxtags.UnaryServerInterceptor(),
		promInterceptor.UnaryServer(),
		grpc_logrus.UnaryServerInterceptor(log.NewEntry(log.New())),
		grpc_validator.UnaryServerInterceptor(),
		grpc_auth.UnaryServerInterceptor(auth.BasicAuthFunc()),
		grpc_recovery.UnaryServerInterceptor(),
	)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			promInterceptor.StreamServer(),
			grpc_validator.StreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(auth.BasicAuthFunc()),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.StatsHandler(promInterceptor),
	)
	s := &Server{
		server:     server,
		router:     echo.New(),
		db:         db,
		hTTPClient: http.DefaultClient,
		logger:     log.New(),
		config:     cfg,
	}
	s.router.Use(
		middleware.Recover(),
	)
	s.router.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	s.hTTPClient.Timeout = 5 * time.Second
	return s, nil
}

func (s *Server) Run() {
	lis, err := net.Listen("tcp", config.Config.GetString("USERDB_PORT"))
	if err != nil {
		s.router.Logger.Fatal(err.Error())
	}
	defer lis.Close()
	defer s.GetDB().Close()

	mux := cmux.New(lis)
	gMux := mux.Match(cmux.HTTP2())
	hMux := mux.Match(cmux.Any())

	fmt.Printf("starting grpc and http server on port %s\n", config.Config.GetString("USERDB_PORT"))

	egp, _ := errgroup.WithContext(context.Background())
	egp.Go(func() error {
		for {
			time.Sleep(config.Config.GetDuration("USERDB_GC_INTERVAL"))
			s.db.RunValueLogGC(0.7)
		}
	})
	egp.Go(func() error {
		return s.router.Server.Serve(hMux)
	})
	egp.Go(func() error {
		return s.server.Serve(gMux)
	})
	egp.Go(func() error {
		return mux.Serve()
	})
	if err := egp.Wait(); err != nil {
		s.router.Logger.Fatal(err.Error())
	}
}

func (s *Server) Setup(fn func(s *Server) error) {
	if err := fn(s); err != nil {
		s.GetLogger().Fatal(err.Error())
	}
}
