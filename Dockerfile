FROM golang:1.13.6-alpine3.10 as builder
LABEL maintainer="andy.lo-a-foe@philips.com"
RUN apk add --no-cache git openssh gcc musl-dev
WORKDIR /siderite
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# Build
COPY . .
RUN go build .

FROM alpine:latest
RUN apk update && apk add ca-certificates && apk add postgresql-client && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /siderite/siderite /app
ENTRYPOINT ["/app/siderite"]
