FROM golang:1.23-alpine AS build_base
RUN apk add --no-cache git
WORKDIR /src
ARG APP_VERSION

COPY . /src
WORKDIR /src
RUN go mod download
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w -X 'main.version=$APP_VERSION'" -o main .

FROM alpine:3.20
RUN apk add ca-certificates
COPY --from=build_base /src/main /app/main

ENTRYPOINT ["/app/main"]
