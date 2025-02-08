FROM golang:1.23.4-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o shortener-app cmd/main.go

FROM alpine AS runner

RUN apk add --no-cache curl

WORKDIR /app

COPY --from=build /build/shortener-app ./
COPY --from=build /build/config ./config/

CMD ["./shortener-app"]