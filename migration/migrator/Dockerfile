FROM golang:1.23.4-alpine AS build

RUN go install github.com/rubenv/sql-migrate/...@latest

FROM alpine AS runner

COPY --from=build /go/bin/sql-migrate /usr/local/bin/sql-migrate

COPY /user_backend/migrations ./migrations

ENTRYPOINT ["/usr/local/bin/sql-migrate"]