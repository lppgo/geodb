package stream

import (
	"context"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/gofrs/uuid"
	"sync"
)

var objectChan = make(chan *api.UserDetail, 5000)

type Hub struct {
	objectClients map[string]chan *api.UserDetail
	objMu         *sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		objectClients: map[string]chan *api.UserDetail{},
		objMu:         &sync.Mutex{},
	}
}

func (h *Hub) StartObjectStream(ctx context.Context) error {
	for {
		select {
		case obj := <-objectChan:
			if h.objectClients == nil {
				h.objectClients = map[string]chan *api.UserDetail{}
			}

			for _, channel := range h.objectClients {
				if channel != nil {
					channel <- obj
				}
			}
		case <-ctx.Done():
			break
		}
	}
}

func (h *Hub) AddObjectStreamClient(clientID string) string {
	h.objMu.Lock()
	defer h.objMu.Unlock()
	if h.objectClients == nil {
		h.objectClients = map[string]chan *api.UserDetail{}
	}
	if clientID == "" {
		id, _ := uuid.NewV4()
		clientID = id.String()
	}
	h.objectClients[clientID] = make(chan *api.UserDetail)
	return clientID
}

func (h *Hub) RemoveObjectStreamClient(id string) {
	h.objMu.Lock()
	defer h.objMu.Unlock()
	if _, ok := h.objectClients[id]; ok {
		close(h.objectClients[id])
		delete(h.objectClients, id)
	}
}

func (h *Hub) GetClientObjectStream(id string) chan *api.UserDetail {
	h.objMu.Lock()
	defer h.objMu.Unlock()
	if _, ok := h.objectClients[id]; ok {
		return h.objectClients[id]
	}
	return nil
}

func PublishObject(obj *api.UserDetail) {
	objectChan <- obj
}

func (h *Hub) PublishObject(obj *api.UserDetail) {
	PublishObject(obj)
}
