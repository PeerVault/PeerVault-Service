build:
	cd src && go build -o ../bin/peervault main.go

test:
	cd src && go test -v -coverprofile=/tmp/profile.out github.com/Power-LAB/PeerVault/crypto
