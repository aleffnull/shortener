EXE = bin/shortener
TEST_EXE = shortenertest
PORT = 8842
FILE_STORAGE = /tmp/shortener.jsonl
DATABASE_CONN_STRING = postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable
BUILD_VERSION = v1.0.0
BUILD_DATE = $(shell date +'%F %T %Z')
BUILD_COMMIT = $(shell git rev-parse HEAD)

clean:
	rm -rf bin/

build:
	go build -v -o=$(EXE) ./cmd/shortener/...

build_linter:
	go build -v -o=bin/linter ./cmd/linter/...

build_resetter:
	go build -v -o=bin/resetter ./cmd/resetter/...

run:
	go run -ldflags " \
	    -X main.BuildVersion=$(BUILD_VERSION) \
	    -X 'main.BuildDate=$(BUILD_DATE)' \
	    -X main.BuildCommit=$(BUILD_COMMIT) \
	  " ./cmd/shortener/...

mock:
	mockgen -source internal/app/app.go -destination internal/pkg/mocks/mock_app.go -package mocks
	mockgen -source internal/service/delete_url_service.go -destination internal/pkg/mocks/mock_delete_url_service.go -package mocks
	mockgen -source internal/service/audit_service.go -destination internal/pkg/mocks/mock_audit_service.go -package mocks
	mockgen -source internal/service/authorization_service.go -destination internal/pkg/mocks/mock_authorization_service.go -package mocks
	mockgen -source internal/repository/connection.go -destination internal/pkg/mocks/mock_connection.go -package mocks
	mockgen -source internal/pkg/logger/logger.go -destination internal/pkg/mocks/mock_logger.go -package mocks
	mockgen -source internal/pkg/parameters/app_parameters.go -destination internal/pkg/mocks/mock_app_parameters.go -package mocks
	mockgen -source internal/pkg/store/store.go -destination internal/pkg/mocks/mock_store.go -package mocks

unittest:
	go test -v -cover ./...

statictest:
	go vet -vettool=$(shell which statictest) ./...

autotest: build
	$(TEST_EXE) -test.v -test.run=^TestIteration1$$  -binary-path=$(EXE)
	$(TEST_EXE) -test.v -test.run=^TestIteration2$$  -source-path=.
	$(TEST_EXE) -test.v -test.run=^TestIteration3$$  -source-path=.
	$(TEST_EXE) -test.v -test.run=^TestIteration4$$  -binary-path=$(EXE) -server-port=$(PORT)
	$(TEST_EXE) -test.v -test.run=^TestIteration5$$  -binary-path=$(EXE) -server-port=$(PORT)
	$(TEST_EXE) -test.v -test.run=^TestIteration6$$  -source-path=.
	$(TEST_EXE) -test.v -test.run=^TestIteration7$$  -binary-path=$(EXE) -source-path=.
	$(TEST_EXE) -test.v -test.run=^TestIteration8$$  -binary-path=$(EXE)
	$(TEST_EXE) -test.v -test.run=^TestIteration9$$  -binary-path=$(EXE) -source-path=. -file-storage-path=$(FILE_STORAGE)
	$(TEST_EXE) -test.v -test.run=^TestIteration10$$ -binary-path=$(EXE) -source-path=. -database-dsn=$(DATABASE_CONN_STRING)
	$(TEST_EXE) -test.v -test.run=^TestIteration11$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)
	$(TEST_EXE) -test.v -test.run=^TestIteration12$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)
	$(TEST_EXE) -test.v -test.run=^TestIteration13$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)
	$(TEST_EXE) -test.v -test.run=^TestIteration14$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)
	$(TEST_EXE) -test.v -test.run=^TestIteration15$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)

test: unittest statictest autotest

bench:
	go test -bench=. ./...

format:
	find . -name \*.go -exec goimports -v -local "github.com/aleffnull/shortener" -w {} \;

docs:
	godoc -v -play -http=:9090

calc_coverage:
	go test ./... -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "mock" | grep -v ".pb.go" > cover.out
	rm cover.out.tmp

coverage: calc_coverage
	go tool cover -func cover.out

coverage_html: coverage
	go tool cover -html=cover.out

lint: build_linter
	bin/linter ./...

generate: build_resetter
	bin/resetter .

proto:
	mkdir -p internal/pkg/pb
	protoc \
		--go_out=internal/pkg/pb \
		--go-grpc_out=internal/pkg/pb \
		api/shortener/shortener.proto
