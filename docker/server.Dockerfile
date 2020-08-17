FROM golang:1.15-alpine3.12

RUN apk add --no-cache git make gcc libc-dev

ENV GOPATH=""

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make build

# TODO
# Copy to fresh image to keep final image small and secure.
#FROM scratch
