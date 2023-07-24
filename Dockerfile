FROM golang:1.20.4-alpine3.17 AS builder

RUN apk add --no-cache git openssh-client gcc libc-dev

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o runner ./cmd/runner/*.go

FROM alpine:3.17

WORKDIR /app

COPY --from=builder /app/runner .

EXPOSE 8080

ENTRYPOINT ["/app/runner"]
