FROM golang:1.22-alpine3.19 AS builder
RUN apk add --no-cache git
WORKDIR /go/src/github.com/pokt-foundation

COPY . /go/src/github.com/pokt-foundation/portal-http-db/v2

WORKDIR /go/src/github.com/pokt-foundation/portal-http-db/v2
RUN CGO_ENABLED=0 GOOS=linux go build -a -o bin ./main.go

FROM alpine:3.19
WORKDIR /app

ARG IMAGE_TAG
ENV IMAGE_TAG=${IMAGE_TAG}

COPY --from=builder /go/src/github.com/pokt-foundation/portal-http-db/v2/bin ./

ENTRYPOINT ["/app/bin"]
