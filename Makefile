# app

.PHONY: run

run:
	go run cmd/app/main.go


# migrations
# go install github.com/mikefarah/yq/v4@latest
# go install github.com/pressly/goose/v3/cmd/goose@latest

DBSTRING := $(shell yq e '.db.uri' config.yml)
GOOSE := goose -dir migrations postgres "$(DBSTRING)"

.PHONY: goose

goose:
	$(GOOSE) $(filter-out $@,$(MAKECMDGOALS))

%:
	@: