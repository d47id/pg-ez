FROM golang:1.17-alpine3.15 AS build
WORKDIR /go/src/app
COPY go.mod go.sum /go/src/app/
RUN go mod download
COPY backup/*.go /go/src/app/
RUN go build -ldflags "-s -w" -o /go/bin/backup

FROM alpine:3.15
RUN apk --update --no-cache add postgresql-client \
    && mkdir -p /backup
COPY --from=build --chown=postgres:postgres /go/bin/backup /backup
USER postgres
ENTRYPOINT [ "/backup/backup" ]