all: gen test revive vet build

clean:
	@rm -f aq.db*
	@rm -rf bin

build: aq

aq:
	@cd cmd/aq && go build -o ../../bin/aq

vet:
	@go vet ./...

revive:
	@revive ./...

test:
	@go test ./...

test_verbose:
	@go test ./... -v

gen:
	@buf generate

proto-image:
	@buf build -o proto/aq.json

dep:
	@go get -u github.com/bufbuild/buf/cmd/buf github.com/mgechev/revive
	@go mod tidy

