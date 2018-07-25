#! /bin/bash
cd ../
echo "Cloning authenticate.git"
git clone ssh://ci@gitmain.ibfd.org:2222/git/git/repo/authenticate.git
echo "Cloning session.git"
git clone ssh://ci@gitmain.ibfd.org:2222/git/git/repo/session.git
echo "Go getting stuff"
go get github.com/gorilla/mux github.com/denisenkom/go-mssqldb github.com/go-redis/redis github.com/rs/cors
echo "Building server"
cd annotate
# Regular build
#go build -a server/main.go
# Alpine build
CGO_ENABLED=0 go build -a -installsuffix cgo server/main.go
