APP_NAME=slackroom

default: bin

bin:
	mkdir -p bin
	godep go build -i -o ./bin/$(APP_NAME)

test:
	godep go test -timeout=10s -v

format:
	git ls-files | grep '.go$$' | xargs gofmt -w

deps:
	go get github.com/tools/godep
	GOPATH=`pwd`/Godeps/_workspace godep restore

.PHONY: default bin test format deps
