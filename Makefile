EXE = bin/shortener
PORT = 8842
FILE_STORAGE = /tmp/shortener.jsonl
DATABASE_CONN_STRING = postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable

clean:
	rm -rf bin/

build:
	go build -v -o=$(EXE) ./cmd/shortener/...

run:
	go run ./cmd/shortener/...

mock:
	mockgen -source internal/app/app.go -destination internal/pkg/mocks/mock_app.go -package mocks
	mockgen -source internal/pkg/database/connection.go -destination internal/pkg/mocks/mock_connection.go -package mocks
	mockgen -source internal/pkg/logger/logger.go -destination internal/pkg/mocks/mock_logger.go -package mocks
	mockgen -source internal/pkg/parameters/app_parameters.go -destination internal/pkg/mocks/mock_app_parameters.go -package mocks
	mockgen -source internal/pkg/store/store.go -destination internal/pkg/mocks/mock_store.go -package mocks

unittest:
	go test -v -cover ./...

statictest:
	go vet -vettool=$(shell which statictest) ./...

autotest: build
	shortenertestbeta -test.v -test.run=^TestIteration1$$  -binary-path=$(EXE)
	shortenertestbeta -test.v -test.run=^TestIteration2$$  -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration3$$  -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration4$$  -binary-path=$(EXE) -server-port=$(PORT)
	shortenertestbeta -test.v -test.run=^TestIteration5$$  -binary-path=$(EXE) -server-port=$(PORT)
	shortenertestbeta -test.v -test.run=^TestIteration6$$  -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration7$$  -binary-path=$(EXE) -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration8$$  -binary-path=$(EXE)
	shortenertestbeta -test.v -test.run=^TestIteration9$$  -binary-path=$(EXE) -source-path=. -file-storage-path=$(FILE_STORAGE)
	shortenertestbeta -test.v -test.run=^TestIteration10$$ -binary-path=$(EXE) -source-path=. -database-dsn=$(DATABASE_CONN_STRING)
	shortenertestbeta -test.v -test.run=^TestIteration11$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)
	shortenertestbeta -test.v -test.run=^TestIteration12$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)
	shortenertestbeta -test.v -test.run=^TestIteration13$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)
	shortenertestbeta -test.v -test.run=^TestIteration14$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)
	shortenertestbeta -test.v -test.run=^TestIteration15$$ -binary-path=$(EXE) -database-dsn=$(DATABASE_CONN_STRING)

test: unittest statictest autotest
