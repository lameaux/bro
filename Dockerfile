FROM golang:1.22-alpine AS builder
ARG GIT_HASH

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags "-X main.GitHash=$GIT_HASH" -o bro ./cmd/bro/*.go
RUN go build -ldflags "-X main.GitHash=$GIT_HASH" -o brod ./cmd/brod/*.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /bro .
COPY --from=builder /brod .

ENTRYPOINT ["/bro"]
