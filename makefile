COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# VERSION := $(shell docker run --rm -v "$$(pwd):/repo" gittools/gitversion:6.0.5 /repo /output json /showvariable FullSemVer)
VERSION := $(shell gitversion /showvariable FullSemVer)
include deploy/.env
export
run:
	go run -ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'" cmd/go-read-burn/main.go

up:
	docker compose --project-directory deploy up --build --remove-orphans

version:
	@echo $(VERSION)
