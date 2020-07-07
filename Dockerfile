FROM golang:1.14.4-stretch as builder
LABEL maintainer="andy.lo-a-foe@philips.com"
WORKDIR /siderite
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# Build
COPY . .
RUN git rev-parse --short HEAD
RUN GIT_COMMIT=$(git rev-parse --short HEAD) && \
	go build -ldflags "-X siderite/cmd.GitCommit=${GIT_COMMIT}"

FROM golang:1.14.4-stretch
WORKDIR /app
COPY --from=builder /siderite/siderite /app
ENTRYPOINT ["/app/siderite","runner"]
