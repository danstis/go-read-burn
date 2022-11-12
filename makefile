VERSION := $(shell docker run --rm -v "$$(pwd):/repo" gittools/gitversion:5.11.1 /repo /output json /showvariable FullSemVer)
include deploy/.env
export
run:
	go run -ldflags "-s -w -X 'github.com/danstis/go-read-burn/internal/version.Version=$(VERSION)'" cmd/go-read-burn/main.go

up:
	docker compose --project-directory deploy up --build --remove-orphans

version:
	@echo $(VERSION)
