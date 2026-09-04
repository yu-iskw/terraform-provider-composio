default: test

.PHONY: test
test:
	unset TF_ACC && cd "internal/" && go test -count=1 -v ./...

# Pin Acc Terraform via TF_ACC_TERRAFORM_VERSION (from .terraform-version).
# terraform-plugin-testing runs the CLI from a temp dir; tfenv shims then fall
# back to the global default (often below the supported floor) and skip Acc cases.
TF_ACC_TERRAFORM_VERSION ?= $(shell cat .terraform-version 2>/dev/null)

.PHONY: testacc
testacc:
	TF_ACC=1 TF_ACC_TERRAFORM_VERSION="$(TF_ACC_TERRAFORM_VERSION)" go test ./internal/provider/... -v $(TESTARGS) -timeout 120m

.PHONY: clean
clean:
	go clean -cache -modcache -i -r

.PHONY: build
build: gen-docs go-tidy gosec deadcode
	go build -v ./

.PHONY: gosec
gosec:
	gosec ./internal/...

.PHONY: deadcode
deadcode:
	go run golang.org/x/tools/cmd/deadcode -test ./...

.PHONY: upgrade-go-mod
upgrade-go-mod:
	go get -u ./...
	go mod tidy
	go mod vendor

.PHONY: lint
lint: run-trunk-check deadcode gosec

.PHONY: run-trunk-check
run-trunk-check:
	trunk check --all -y

.PHONY: format
format: format-go format-trunk

.PHONY: format-trunk
format-trunk:
	trunk fmt --all

.PHONY: format-go
format-go:
	go fmt ./internal/...

.PHONY: install
install:
	go build -v ./ && go install .

.PHONY: gen-docs
gen-docs:
	go generate ./...

.PHONY: go-tidy
go-tidy:
	go mod tidy

.PHONY: setup-dev
setup-dev: unset-git-hooks setup-trunk

.PHONY: unset-git-hooks
unset-git-hooks:
	git config --unset-all core.hooksPath || true

.PHONY: setup-trunk
setup-trunk:
	trunk git-hooks sync

.PHONY: update
update: update-trunk

.PHONY: update-trunk
update-trunk:
	trunk upgrade
