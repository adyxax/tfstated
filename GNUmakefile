SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.ONESHELL:
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
.DEFAULT_GOAL := help

##### Quality ##################################################################
.PHONY: check
check: ## run all code checks
	go mod verify
	go vet ./...
	go test -race -buildvcs -vet=off ./...

.PHONY: cover
cover: ## compute tests coverage
	go test -cover -coverprofile cover.out -race -buildvcs -vet=off ./...
	go tool cover -html=cover.out

.PHONY: tidy
tidy: ## tidy up the code
	go fmt ./...
	go mod tidy -v

##### Development ##############################################################
.PHONY: build
build: ## build the code
	go build -o ./tfstated ./cmd/tfstated/

.PHONY: clean
clean: ## clean the code
	rm -f ./tfstated

.PHONY: run
run: ## run the code
	go run ./cmd/tfstated | jq

##### Operations ###############################################################
.PHONY: push
push: tidy no-dirty check ## push changes to git remote
	git push git main

.PHONY: deploy
deploy: build ## deploy changes to the production server
	umask 077
	if [ -n "$${SSH_PRIVATE_KEY:-}" ]; then
	    cleanup() {
	        rm -f private_key
	    }
	    trap cleanup EXIT
	    printf '%s' "$$SSH_PRIVATE_KEY" | base64 -d > private_key
	    SSHOPTS="-i private_key -o StrictHostKeyChecking=accept-new"
	fi
	rsync -e "ssh $${SSHOPTS:-}" ./tfstated tfstated@tfstated.adyxax.org:
	ssh $${SSHOPTS:-} tfstated@tfstated.adyxax.org "chmod +x tfstated; systemctl --user restart tfstated"

##### Utils ####################################################################
.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: help
help:
	@grep -E '^[a-zA-Z\/_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' \
	| sort

.PHONY: no-dirty
no-dirty:
	git diff --exit-code
