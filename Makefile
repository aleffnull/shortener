clean:
	rm -rf bin/

build:
	go build -v -o=bin/shortener ./cmd/shortener/...

run:
	go run ./cmd/shortener/...

mock:
	mockgen -source internal/app/app.go -destination internal/pkg/mocks/mock_app.go -package mocks
	mockgen -source internal/store/store.go -destination internal/pkg/mocks/mock_store.go -package mocks
	mockgen -source internal/pkg/logger/logger.go -destination internal/pkg/mocks/mock_logger.go -package mocks

unittest:
	go test -v -cover ./...

statictest:
	go vet -vettool=$(shell which statictest) ./...

autotest: build
	shortenertestbeta -test.v -test.run=^TestIteration1$$ -binary-path=bin/shortener
	shortenertestbeta -test.v -test.run=^TestIteration2$$ -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration3$$ -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration4$$ -binary-path=bin/shortener -server-port=8842
	shortenertestbeta -test.v -test.run=^TestIteration5$$ -binary-path=bin/shortener -server-port=8842
	shortenertestbeta -test.v -test.run=^TestIteration6$$ -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration7$$ -binary-path=bin/shortener -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration8$$ -binary-path=bin/shortener
	shortenertestbeta -test.v -test.run=^TestIteration9$$ -binary-path=bin/shortener -source-path=. -file-storage-path=/tmp/shortener.jsonl

test: unittest statictest autotest
