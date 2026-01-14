.PHONY: tools generate generate-identity generate-dummy fmt identity-run dummy-run migrate-identity migrate-dummy tidy

TOOLS = \
	goa.design/goa/v3/cmd/goa@latest \
	github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
	github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
	github.com/air-verse/air@latest

all: generate

tools:
	@for tool in $(TOOLS); do \
		GO111MODULE=on go install $$tool; \
	done

generate: generate-identity generate-dummy

generate-identity:
	cd identity-api && goa gen github.com/vidwadeseram/go-boilerplate/identity-api/design
	cd identity-api && sqlc generate

generate-dummy:
	cd dummy-api && goa gen github.com/vidwadeseram/go-boilerplate/dummy-api/design
	cd dummy-api && sqlc generate

fmt:
	cd identity-api && gofmt -w $$(rg --files -g'*.go')
	cd dummy-api && gofmt -w $$(rg --files -g'*.go')

identity-run:
	cd identity-api && go run ./cmd/identity-api serve

dummy-run:
	cd dummy-api && go run ./cmd/dummy-api serve

migrate-identity:
	cd identity-api && go run ./cmd/identity-api migrate --dir migrations --action up

migrate-dummy:
	cd dummy-api && go run ./cmd/dummy-api migrate --dir migrations --action up

tidy:
	cd identity-api && go mod tidy
	cd dummy-api && go mod tidy
