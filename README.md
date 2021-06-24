# cobbles-api

## current antipatterns
- do not create a `models/` package to correctly store types belonging to data. This is to prevent unessecary creep of work that happens with passionate development.

## useful commands
- `docker-compose up -d`
- `fresh` (requires fresh)
- `migrate create -ext sql -dir ./migrations "$NAME"`
- `migrate -path ./migrations -database "$DATABASE" up`

## development & auto-reload

- Install `fresh`: 

    `go get github.com/pilu/fresh`

- create run.sh
    ```
    export DATABASE_URL="postgres://xxx"
    export SERVER_SECRET=xxx
    export PORT=8080
    export AWS_ACCESS_KEY_ID=xxxx
    export AWS_SECRET_ACCESS_KEY=xxx
    export SNS_APP_ARN=xxx

    fresh
    ```

- `chmod 755 run.sh`
- `./run.sh`

## deployment
examples are for `cobbles-api-dev` development app. please adapt appropriately.
- maybe read https://devcenter.heroku.com/articles/build-docker-images-heroku-yml
- acknowledge `heroku.yml`
- add the respective heroku remote
- `git push heroku master`