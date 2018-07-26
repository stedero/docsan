#! /bin/bash
# Alpine build
CGO_ENABLED=0 go build -a -installsuffix cgo docsan.go version.go
