format:
	goimports -local "github.com/LasTshaMAN/streaming" -w ./
	# We need to run `gofmt` with `-s` flag as well (best practice, linters require it).
	# `goimports` doesn't support `-s` flag just yet.
	# For details see https://github.com/golang/go/issues/21476
	gofmt -w -s ./

lint_docker:
	docker run --rm -v $(GOPATH)/pkg/mod:/go/pkg/mod:ro -v `pwd`:/`pwd`:ro -w /`pwd` golangci/golangci-lint:v1.27-alpine golangci-lint run --deadline=5m -v

test:
	go test ./... -race -timeout 60s -count=1 -cover -coverprofile=coverage.txt && go tool cover -func=coverage.txt

# TODO
mock_gen:

client_run:
	go run ./cmd/client

# TODO
# Rewrite this with docker-compose
# TODO
# (and check how that affects throughput)
server_run_docker:
	go run ./cmd/server

proto_gen:
	protoc \
	--go_out=:. \
	--go-grpc_out=:. \
	./streaming.proto

# Spin up all the dependencies.
deps_up:
	docker-compose -f docker/docker-compose.yml up -d --build

# Shut down all the dependencies.
deps_down:
	docker-compose -f docker/docker-compose.yml down
