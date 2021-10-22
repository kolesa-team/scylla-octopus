FROM golang:1.17 as build

WORKDIR /code

ARG GIT_REV
ARG DATE
ARG VERSION

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .

RUN DATE=$DATE GIT_REV=$GIT_REV VERSION=$VERSION make build

FROM alpine:3.13

COPY --from=build /code/dist/scylla-octopus /usr/local/bin/

ADD config /config

ENTRYPOINT ["/usr/local/bin/scylla-octopus"]
