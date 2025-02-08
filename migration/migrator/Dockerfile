FROM golang:1.23.4-alpine AS build

RUN go install github.com/rubenv/sql-migrate/...@latest

FROM alpine AS runner

WORKDIR /app

COPY --from=build /go/bin/sql-migrate /usr/local/bin/sql-migrate

COPY /migration ./migration

ENTRYPOINT ["/usr/local/bin/sql-migrate"]