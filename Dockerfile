FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/github.com/pokt-foundation

COPY . /go/src/github.com/pokt-foundation/portal-http-db

WORKDIR /go/src/github.com/pokt-foundation/portal-http-db
RUN CGO_ENABLED=0 GOOS=linux go build -a -o bin ./main.go

FROM alpine:3.16.0
WORKDIR /app

ARG IMAGE_TAG
ENV IMAGE_TAG=${IMAGE_TAG}

COPY --from=builder /go/src/github.com/pokt-foundation/portal-http-db/v2/bin ./

ENTRYPOINT ["/app/bin"]
