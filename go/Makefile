# Copyright 2026 The kpt Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GOLANGCI_LINT_VERSION ?= 2.12.2

.PHONY: all
all: fix vet fmt test lint

MODULES = $(shell find . -name 'go.mod' -print)

.PHONY: fix
fix: $(MODULES)
	@for f in $(^D); do (cd $$f; echo "Fixing $$f"; go fix -omitzero=false ./...) || exit 1; done

.PHONY: fmt
fmt: $(MODULES)
	@for f in $(^D); do (cd $$f; echo "Formatting $$f"; go fmt ./...); done

.PHONY: lint
lint: $(MODULES)
	@if ! command -v golangci-lint >/dev/null 2>&1 || \
	  ! golangci-lint version --short 2>/dev/null | grep -qx "$(GOLANGCI_LINT_VERSION)"; then \
		echo "Installing golangci-lint v$(GOLANGCI_LINT_VERSION)..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION); \
	fi
	@for f in $(^D); do \
	(cd $$f; echo "Linting $$f"; golangci-lint run ./...) || exit 1; \
	done

.PHONY: test
test: $(MODULES)
	@for f in $(^D); do (cd $$f; echo "Testing $$f"; go test -v ./...) || exit 1; done

.PHONY: vet
vet: $(MODULES)
	@#for f in $(^D); do (cd $$f; echo "Checking $$f"; go run honnef.co/go/tools/cmd/staticcheck@latest ./...); done
	@for f in $(^D); do (cd $$f; echo "Vetting $$f"; go vet ./...); done
