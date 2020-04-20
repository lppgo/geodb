package main

import (
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/autom8ter/userdb/server"
	"github.com/autom8ter/userdb/services"
	log "github.com/sirupsen/logrus"
)

func main() {
	s, err := server.NewServer()
	if err != nil {
		log.Fatal(err.Error())
	}
	s.Setup(func(server *server.Server) error {
		api.RegisterUserDBServer(s.GetGRPCServer(), services.NewUserDB(s.GetDB(), s.GetStream()))
		return nil
	})
	s.Run()
}
