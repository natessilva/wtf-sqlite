DB_FILES := $(wildcard db/**/*.sql)
TEMPL_FILES := $(wildcard templates/*.templ)

.PHONY: generate run test deploy

generate: .sqlc_generated .templ_generated

.sqlc_generated: $(DB_FILES)
	@sqlc generate
	@touch .sqlc_generated

.templ_generated: $(TEMPL_FILES)
	@templ generate
	@touch .templ_generated

run: generate
	@go run cmd/server/main.go

test: generate
	@go test ./...

deploy: generate
	@scripts/deploy.sh