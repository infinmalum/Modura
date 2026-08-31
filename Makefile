.PHONY: verify generate generated-clean generate-go generate-admin backend-format backend-lint backend-test backend-build admin-verify

verify: generated-clean backend-format backend-lint backend-test backend-build admin-verify

generated-clean:
	@before="$$(find backend/internal/api/generated backend/internal/platform/database/dbgen admin/src/api/generated -type f -print0 | sort -z | xargs -0 sha256sum)"; \
	$(MAKE) generate; \
	after="$$(find backend/internal/api/generated backend/internal/platform/database/dbgen admin/src/api/generated -type f -print0 | sort -z | xargs -0 sha256sum)"; \
	test "$$before" = "$$after"

generate: generate-go generate-admin

generate-go:
	@mkdir -p backend/internal/api/generated
	oapi-codegen -config api/oapi-codegen.yaml -o backend/internal/api/generated/openapi.gen.go api/openapi.yaml
	cd backend && sqlc generate

generate-admin:
	cd admin && npm run generate

backend-format:
	@test -z "$$(gofmt -l backend)"

backend-lint:
	cd backend && GOCACHE="$$PWD/.cache/go-build" GOLANGCI_LINT_CACHE="$$PWD/.cache/golangci-lint" golangci-lint run ./...

backend-test:
	cd backend && GOCACHE="$$PWD/.cache/go-build" go test ./...

backend-build:
	cd backend && GOCACHE="$$PWD/.cache/go-build" go build ./cmd/...

admin-verify:
	cd admin && npm run format:check
	cd admin && npm run lint
	cd admin && npm run typecheck
	cd admin && npm run test
	cd admin && npm run build
