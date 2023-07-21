COMMIT := $(git rev-parse HEAD)
DATE := $(date)
VERSION := $(shell docker run --rm -v "$$(pwd):/repo" gittools/gitversion:5.11.1 /repo /output json /showvariable FullSemVer)
include deploy/.env
export
run:
	go run -ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'" cmd/go-read-burn/main.go

up:
	docker compose --project-directory deploy up --build --remove-orphans

version:
	@echo $(VERSION)
