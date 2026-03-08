# supported versions here: https://hub.docker.com/_/golang
ARG GOLANG_BUILDER_VERSION=alpine3.23
ARG ALPINE_VERSION=3.23

########################
## builder image
########################
FROM golang:${GOLANG_BUILDER_VERSION} AS builder

WORKDIR /src

# copy the source and build the binary
COPY . ./
RUN go mod tidy
RUN go build -o build/blogroll-md ./cmd/blogroll-md/app.go
RUN echo "finished building blogroll-md"

########################
## release image
########################
FROM alpine:${ALPINE_VERSION} AS release

# Import blogroll-md binary from builder
WORKDIR /app
COPY --from=builder /src/build/blogroll-md /app/bin/blogroll-md

# Add non-root user for running blogroll-md
RUN adduser --home /nonexistent --no-create-home --disabled-password blogroll-md
RUN chown -R blogroll-md:blogroll-md /app
USER blogroll-md

# Add container metadata
LABEL org.opencontainers.image.authors="blebon"

CMD ["bin/blogroll-md"]