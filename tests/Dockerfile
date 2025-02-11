FROM golang:1.23.4-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

CMD ["go", "test", "-v", "./tests"]