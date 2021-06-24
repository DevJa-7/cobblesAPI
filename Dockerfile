FROM golang:1.13-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git curl

WORKDIR /go/src/github.com/lambdacollective/cobbles-api

RUN go get -u github.com/kardianos/govendor

COPY ./vendor/vendor.json ./vendor/vendor.json
RUN govendor sync

COPY . .

RUN go install -v ./...

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root

COPY --from=0 /go/bin/cobbles-api .

CMD ["/bin/sh", "-c", "./cobbles-api"]
