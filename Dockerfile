FROM golang:1.23-bookworm AS builder
ARG VERSION
ARG BUILD_HASH
ARG BUILD_DATE

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags "-X main.Version=$VERSION -X main.BuildHash=$BUILD_HASH -X main.BuildDate=$BUILD_DATE" -o bro ./cmd/client/*.go
RUN go build -ldflags "-X main.Version=$VERSION -X main.BuildHash=$BUILD_HASH -X main.BuildDate=$BUILD_DATE" -o brod ./cmd/server/*.go

FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=builder /app/bro .
COPY --from=builder /app/brod .

ENTRYPOINT ["/bro"]
