format:
	goimports -local "github.com/LasTshaMAN/streaming" -w ./
	# We need to run `gofmt` with `-s` flag as well (best practice, linters require it).
	# `goimports` doesn't support `-s` flag just yet.
	# For details see https://github.com/golang/go/issues/21476
	gofmt -w -s ./

build:
	go build -o ./bin/server ./cmd/server/

lint_docker:
	docker run --rm -v $(GOPATH)/pkg/mod:/go/pkg/mod:ro -v `pwd`:/`pwd`:ro -w /`pwd` golangci/golangci-lint:v1.27-alpine golangci-lint run --deadline=5m -v

test:
	go test ./... -race -timeout 60s -count=1 -cover -coverprofile=coverage.txt && go tool cover -func=coverage.txt

# TODO - try https://github.com/matryer/moq is action for mocks
mock_gen:

run_client:
	mkdir -p log
	rm -f ./log/client.log 2>&1
	go run ./cmd/client >> ./log/client.log 2>&1

# TODO - scale server with docker-compose and see how that affects throughput
run_server:
	rm -f ./log/server.log 2>&1
	mkdir -p log
	docker-compose -p server -f docker/docker-compose-server.yml up -d --build
	docker-compose -p server -f docker/docker-compose-server.yml logs -f server >> ./log/server.log 2>&1

stop_server:
	docker-compose -p server -f docker/docker-compose-server.yml down

proto_gen:
	protoc \
	--go_out=:. \
	--go-grpc_out=:. \
	./streaming.proto

# Spin up all the dependencies.
up_deps:
	docker-compose -p deps -f docker/docker-compose-deps.yml up -d

# Shut down all the dependencies.
down_deps:
	docker-compose -p deps -f docker/docker-compose-deps.yml down
