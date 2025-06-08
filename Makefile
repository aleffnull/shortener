clean:
	rm -rf bin/

build:
	go build -v -o=bin/shortener ./cmd/shortener/...

run:
	go run cmd/shortener/main.go

unittest:
	go test -v -cover ./...

statictest:
	go vet -vettool=$(shell which statictest) ./...

autotest: build
	shortenertestbeta -test.v -test.run=^TestIteration1$$ -binary-path=bin/shortener
	shortenertestbeta -test.v -test.run=^TestIteration2$$ -source-path=.
	shortenertestbeta -test.v -test.run=^TestIteration3$$ -source-path=.

test: unittest statictest autotest
