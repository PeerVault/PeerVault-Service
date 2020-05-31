GOVERSION=1.12
DEFAULT_RELAY="/ip4/37.187.1.229/tcp/23003/ipfs/QmeFecyqtgzYx1TFN9vYTroMGNo3DELtDZ63FpjqUd6xfW"

.peervault-linux.db:
	touch .peervault-linux.db

# Code Style
fmt:
	go fmt

lint:
	golint src/...

sanity: fmt lint

build:
	cd src && GOOS=darwin go build -ldflags "-w" -o ../bin/peervault main.go

build-linux:
	cd src && GOOS=linux go build -ldflags "-w" -o ../bin/peervault-linux main.go

run-linux: build-linux .peervault-linux.db
	docker run \
		--rm -it \
		-p 4445:4444 -p 5556:5555 \
		-v $$(pwd)/bin/peervault-linux:/usr/local/bin/peervault-linux \
		-v $$(pwd)/.peervault-linux.db:/var/peervault.db \
		golang:$(GOVERSION) \
		/usr/local/bin/peervault-linux \
			-dev \
			--log 9 \
			--apiAddr 0.0.0.0:4444 \
			--wsAddr 0.0.0.0:5555 \
			--relay "$(DEFAULT_RELAY)" \
			--bbolt /var/peervault.db

run: build
	./bin/peervault -dev --log 9 --relay "$(DEFAULT_RELAY)" --bbolt /Users/pierozi/.peervault/bbolt-dev.db

test:
	cd src && go test -v -coverprofile=/tmp/profile.out github.com/Power-LAB/PeerVault/crypto
