BINARY_NAME_SERVER=pongo-server
BINARY_NAME_CLIENT=pongo-client
BUILD_DIR=./build

.PHONY: build_server run_server clean

build-server:
	go build -o $(BUILD_DIR)/$(BINARY_NAME_SERVER) main.go

run-server:
	go run main.go --port=8080

build-client:
	go build -o $(BUILD_DIR)/$(BINARY_NAME_CLIENT) mock/client.go

run-client:
	go run mock/client.go --server="pongo-server-42917a26f363.herokuapp.com:80" --name=TestPlayer

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)
