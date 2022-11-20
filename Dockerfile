#
## Dependencies Stage
FROM golang:1.19.3-alpine3.15 AS dependencies-stage
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

#
## Build Stage
FROM dependencies-stage AS build-stage
COPY . .
RUN cd yahoonews-server && go build -o newsxu

#
## Build Client Stage
FROM node:9.11.2-alpine AS build-client-stage
WORKDIR /app

COPY yahoonews-server yahoonews-server 
RUN cd yahoonews-server/client && npm install && npm run build

#
## Release Stage
FROM alpine:3.15 AS release-stage
WORKDIR /app/yahoonews-server

COPY data ../data
COPY --from=build-stage /app/yahoonews-server/newsxu newsxu
COPY --from=build-client-stage /app/yahoonews-server/public public

CMD []
ENTRYPOINT [ "/app/yahoonews-server/newsxu" ]