FROM golang:1.20.4-alpine3.17 AS builder

RUN apk add --no-cache git openssh-client gcc libc-dev

# Setup SSH for private repo clone
COPY ./.ssh /root/.ssh
RUN git config --global url.git@github.com:.insteadOf https://github.com/

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o runner ./cmd/runner/*.go

FROM alpine:3.17

WORKDIR /app

COPY --from=builder /app/runner .

EXPOSE 8080

ENTRYPOINT ["/app/runner"]
