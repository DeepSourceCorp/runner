FROM golang:1.20.4-alpine3.17 AS builder

RUN apk add --no-cache git openssh-client gcc libc-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=1.0.0-beta.x
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o runner ./cmd/runner/*.go

FROM alpine:3.17

WORKDIR /app

COPY --from=builder /app/runner .

EXPOSE 8080

ENTRYPOINT ["/app/runner"]
