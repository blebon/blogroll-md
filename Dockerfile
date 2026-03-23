# supported versions here: https://hub.docker.com/_/golang
ARG GOLANG_BUILDER_VERSION=alpine3.23
ARG ALPINE_VERSION=3.23

ARG NAME=blogroll-md

########################
## builder image
########################
FROM golang:${GOLANG_BUILDER_VERSION} AS builder

ARG NAME

WORKDIR /src

# copy the source and build the binary
COPY . ./
RUN go mod tidy
RUN go build -o build/${NAME} ./cmd/blogroll-md/app.go
RUN echo "finished building ${NAME}"

########################
## release image
########################
FROM alpine:${ALPINE_VERSION} AS release

ARG NAME

# Import blogroll-md binary from builder
COPY --from=builder /src/build/${NAME} /usr/local/bin/${NAME}

# Add non-root user for running blogroll-md
RUN adduser --home /nonexistent --no-create-home --disabled-password ${NAME}
USER ${NAME}

WORKDIR /data

# Add container metadata
LABEL org.opencontainers.image.authors="blebon"

CMD ["blogroll-md"]