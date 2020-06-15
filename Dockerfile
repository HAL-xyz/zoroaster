FROM golang:1.12-alpine AS build

RUN apk update && apk upgrade && \
    apk add --no-cache git openssh make build-base

WORKDIR /usr/src/app
COPY go.mod .
COPY go.sum .
RUN go get -d -v ./...

COPY . .
RUN go build -o zoroaster


FROM alpine
COPY --from=build /usr/src/app/zoroaster /bin/zoroaster
ENTRYPOINT ["/bin/zoroaster"]
