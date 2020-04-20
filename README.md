# UserDB- A Persistent Geospatial Database

    go get github.com/autom8ter/userdb
    docker pull colemanword/userdb:latest
    
UserDB is a persistant geospatial database built using [Badger](https://github.com/dgraph-io/badger) gRPC, and the Google Maps API

## Features

- [x] Concurrent ACID transactions
- [x] Real-Time Server-Client Object Geolocation Streaming
- [x] Persistent User Storage
- [x] gRPC Protocol
- [x] Prometheus Metrics (/metrics endpoint)
- [x] User Creation timeseries exposed with Prometheus metrics
- [x] Configurable(12-factor)
- [x] Basic Authentication
- [x] Docker Image
- [x] Sample Docker Compose File
- [ ] Kubernetes Manifests
- [ ] REST Translation Layer
- [ ] Horizontal Scaleability(Raft Protocol)
