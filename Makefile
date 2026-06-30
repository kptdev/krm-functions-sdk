.PHONY: go
go: ## Run all e2e tests
	cd go && $(MAKE) all

# find all subdirectories with a go.mod file in them
GO_MOD_DIRS = $(shell find . -name 'go.mod' -exec sh -c 'echo \"$$(dirname "{}")\" ' \; )
# NOTE: the above line is complicated for Mac and busybox compatibilty reasons.
# It is meant to be equivalent with this:  find . -name 'go.mod' -printf "'%h' " 

.PHONY: tidy
tidy:
	@for f in $(GO_MOD_DIRS); do (cd $$f; echo "Tidying $$f"; go mod tidy) || exit 1; done

.PHONY: test
test:
	make -C go test
