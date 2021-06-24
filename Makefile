POSTGRES_URL := postgres://postgres@localhost:5432?sslmode=disable
TEST_POSTGRES_URL := postgres://postgres@localhost:5432/test?sslmode=disable
STAGE_POSTGRES_URL := postgres://uhbycifvppbdzr:d324585a5d5493cd11050ce34942413e9b94c6ec0aee6e2924c2a080147c136a@ec2-174-129-41-12.compute-1.amazonaws.com:5432/d78s5hm4c1acd1

MIGRATE := $(shell command -v migrate 2> /dev/null)
GOVENDOR := $(shell command -v govendor 2> /dev/null)

migrate-dev := migrate -database ${POSTGRES_URL} -source file://./migrations
migrate-stage := migrate -database ${STAGE_POSTGRES_URL} -source file://./migrations
migrate-test := migrate -database ${TEST_POSTGRES_URL} -source file://./migrations

# Heroku sets DATABASE_URL with CI Postgres
migrate-ci := migrate -database ${DATABASE_URL} -source file://./migrations


test:
	govendor test +local

test-ci: cli-deps db-up-ci
	# Heroku runs govendor sync automatically
	GOCACHE=off govendor test -v +local

cli-deps:
ifndef MIGRATE
	go get -tags 'postgres' -u github.com/golang-migrate/migrate/cmd/migrate
endif

ifndef GOVENDOR
	go get -u github.com/kardianos/govendor
endif

pgcli:
	pgcli ${POSTGRES_URL}

pgcli-test:
	pgcli ${TEST_POSTGRES_URL}

pgcli-stage:
	pgcli ${STAGE_POSTGRES_URL}

gen-token:
	SERVER_SECRET=test go run ./cmd/gen-token/main.go

gen-token-stage:
	SERVER_SECRET=lolnotsecretserversecret go run ./cmd/gen-token/main.go

serve: db-up
	DATABASE_URL=${POSTGRES_URL} SERVER_SECRET=${SERVER_SECRET} go run *.go

db-up:
	$(migrate-dev) up
	$(migrate-test) up

db-up-stage:
	$(migrate-stage) up

db-up-ci:
	$(migrate-ci) up

db-drop:
	$(migrate-dev) drop
	$(migrate-test) drop

new-migration:
	$(migrate-dev) create -ext sql -dir ./migrations ${NAME}
