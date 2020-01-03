build:
	cd src && go build -ldflags "-w" -o ../bin/peervault main.go

run: build
	./bin/peervault -dev --log 9 --relay "/ip4/37.187.1.229/tcp/23003/ipfs/QmeFecyqtgzYx1TFN9vYTroMGNo3DELtDZ63FpjqUd6xfW" --bbolt /Users/pierozi/.peervault/bbolt-dev.db

test:
	cd src && go test -v -coverprofile=/tmp/profile.out github.com/Power-LAB/PeerVault/crypto
