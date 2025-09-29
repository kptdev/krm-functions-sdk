.PHONY: go
go: ## Run all e2e tests
	cd go && $(MAKE) all

.PHONY: go-ci
go-ci: 
	cd go && $(MAKE) ci
