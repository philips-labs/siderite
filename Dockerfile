FROM golang:1.16.4-alpine3.13 as builder
RUN apk add --no-cache git
WORKDIR /siderite
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# Build
COPY . .
RUN git rev-parse --short HEAD
RUN GIT_COMMIT=$(git rev-parse --short HEAD) && \
	go build -ldflags "-X siderite/cmd.GitCommit=${GIT_COMMIT}" -o siderite cli/main.go

FROM golang:1.16.4-alpine3.13
LABEL maintainer="andy.lo-a-foe@philips.com"
WORKDIR /app
COPY --from=builder /siderite/siderite /app
ENTRYPOINT ["/app/siderite","runner"]
