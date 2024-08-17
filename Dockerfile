FROM golang:1.22-alpine AS builder
ARG GIT_HASH

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-X main.GitHash=$GIT_HASH" -o bro ./**/*.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /bro .

ENTRYPOINT ["/bro"]
