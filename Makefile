clean:
	rm -rf bin/

build:
	go build -v -o=bin/shortener ./cmd/shortener/...

statictest:
	go vet -vettool=$(shell which statictest) ./...

autotest:
	shortenertestbeta -test.v -test.run=^TestIteration1$$ -binary-path=bin/shortener

test: statictest autotest
