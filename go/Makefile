.PHONY: all
all: fix vet fmt test lint

GOPATH := $(shell go env GOPATH)
GOBIN := $(shell go env GOPATH)/bin
OUT_DIR := .out
MODULES = $(shell find . -name 'go.mod' -print)

.PHONY: api
api:
	(which swagger || go install github.com/go-swagger/go-swagger/cmd/swagger@v0.27.0)
	$(GOPATH)/bin/swagger generate spec -m -w api/kptfile/v1 -o ../openapi/kptfile.yaml

.PHONY: fix
fix: $(MODULES)
	@for f in $(^D); do (cd $$f; echo "Fixing $$f"; go fix ./...) || exit 1; done

.PHONY: fmt
fmt: $(MODULES)
	@for f in $(^D); do (cd $$f; echo "Formatting $$f"; go fmt ./...); done

.PHONY: lint
lint: $(MODULES)
	@for f in $(^D); do \
	(cd $$f; echo "Checking golangci-lint $$f"; \
	(which $(GOPATH)/bin/golangci-lint || go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0); \
	$(GOPATH)/bin/golangci-lint run ./...); \
	done

.PHONY: test
test: $(MODULES)
	@for f in $(^D); do (cd $$f; echo "Testing $$f"; go test ./...) || exit 1; done

.PHONY: vet
vet: $(MODULES)
	@#for f in $(^D); do (cd $$f; echo "Checking $$f"; go run honnef.co/go/tools/cmd/staticcheck@latest ./...); done
	@for f in $(^D); do (cd $$f; echo "Vetting $$f"; go vet ./...); done