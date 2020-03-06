FROM golang:1.14-alpine3.11 as builder
LABEL maintainer="andy.lo-a-foe@philips.com"
RUN apk add --no-cache git openssh gcc musl-dev
WORKDIR /siderite
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# Build
COPY . .
RUN git rev-parse --short HEAD
RUN GIT_COMMIT=$(git rev-parse --short HEAD) && \
	go build -ldflags "-X github.com/philips-labs/siderite/cmd.GitCommit=${GIT_COMMIT}"

FROM alpine:latest
RUN apk update && apk add ca-certificates && apk add postgresql-client && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /siderite/siderite /app
ENTRYPOINT ["/app/siderite","runner"]
