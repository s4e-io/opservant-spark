BINARY   := spark
CMD      := ./cmd/opservant-spark
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -s -w -X main.version=$(VERSION)

.PHONY: spark \
	spark-linux-amd64 spark-linux-arm64 \
	spark-darwin-amd64 spark-darwin-arm64 \
	spark-windows-amd64 spark-windows-arm64 \
	spark-all \
	test vet clean

spark:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) $(CMD)

spark-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-amd64 $(CMD)

spark-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-arm64 $(CMD)

spark-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-darwin-amd64 $(CMD)

spark-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-darwin-arm64 $(CMD)

spark-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-windows-amd64.exe $(CMD)

spark-windows-arm64:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-windows-arm64.exe $(CMD)

spark-all: spark-linux-amd64 spark-linux-arm64 spark-darwin-amd64 spark-darwin-arm64 spark-windows-amd64 spark-windows-arm64

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY) $(BINARY)-* $(BINARY)*.exe
