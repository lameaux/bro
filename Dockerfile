FROM golang:1.23-bookworm AS builder
ARG GIT_HASH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -ldflags "-X main.GitHash=$GIT_HASH" -o bro ./cmd/client/*.go
RUN go build -ldflags "-X main.GitHash=$GIT_HASH" -o brod ./cmd/server/*.go

FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=builder /app/bro .
COPY --from=builder /app/brod .

ENTRYPOINT ["/bro"]
