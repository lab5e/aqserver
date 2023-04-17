default: aq listen

all: gen test revive vet build

clean:
	@rm -f aq.db*
	@rm -rf bin
	@rm -rf logs

build: aq

aq:
	@cd cmd/$@ && go build -o ../../bin/$@

listen:
	@cd cmd/$@ && go build -o ../../bin/$@

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
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install github.com/mgechev/revive@latest
	
count:
	@echo "Linecounts excluding generated and third party code"
	@gocloc --not-match-d='apipb|openapi|third_party' .


