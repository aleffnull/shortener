clean:
	rm -rf bin/

build:
	go build -v -o=bin/shortener ./cmd/shortener/...

unittest:
	go test -v -cover ./...

statictest:
	go vet -vettool=$(shell which statictest) ./...

autotest: build
	shortenertestbeta -test.v -test.run=^TestIteration1$$ -binary-path=bin/shortener
	shortenertestbeta -test.v -test.run=^TestIteration2$$ -source-path=.

test: unittest statictest autotest
