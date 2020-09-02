.PHONY: build
build:
	go build -o mychurnero ./cmd/mychurnero
.PHONY: start-testenv
start-testenv:
	(cd testenv/monero ; bash start.sh)
