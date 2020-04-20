package services

import (
	api "github.com/autom8ter/userdb/gen/go/userdb"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"regexp"
)

func (p *UserDB) Stream(r *api.StreamRequest, ss api.UserDB_StreamServer) error {
	clientID := p.hub.AddObjectStreamClient(r.ClientId)
	for {
		select {
		case msg := <-p.hub.GetClientObjectStream(clientID):
			if len(r.Emails) > 0 {
				if funk.ContainsString(r.Emails, msg.Email) {
					if err := ss.Send(&api.StreamResponse{
						User: msg,
					}); err != nil {
						log.Error(err.Error())
					}
				}
			} else {
				if err := ss.Send(&api.StreamResponse{
					User: msg,
				}); err != nil {
					log.Error(err.Error())
				}
			}
		case <-ss.Context().Done():
			p.hub.RemoveObjectStreamClient(clientID)
			break
		}
	}
}

func (p *UserDB) StreamRegex(r *api.StreamRegexRequest, ss api.UserDB_StreamRegexServer) error {
	clientID := p.hub.AddObjectStreamClient(r.ClientId)
	for {
		select {
		case msg := <-p.hub.GetClientObjectStream(clientID):
			if r.Regex != "" {
				match, err := regexp.MatchString(r.Regex, msg.Email)
				if err != nil {
					return err
				}
				if match {
					if err := ss.Send(&api.StreamRegexResponse{
						User: msg,
					}); err != nil {
						log.Error(err.Error())
					}
				}
			} else {
				if err := ss.Send(&api.StreamRegexResponse{
					User: msg,
				}); err != nil {
					log.Error(err.Error())
				}
			}
		case <-ss.Context().Done():
			p.hub.RemoveObjectStreamClient(clientID)
			break
		}
	}
}
