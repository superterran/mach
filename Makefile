ifeq (run,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## creates binary
	go build .
install: build ## compiles and installs into system
	sudo cp mach /usr/local/bin/mach
	sudo chmod +x /usr/local/bin/mach
run: ## run the command through go, accepts args i.e. `make run -- build -h`
	go run ./main.go $(RUN_ARGS)
test:
	go test --cover ./... 
coverage: ## run test suite suitable for codecov.io
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...