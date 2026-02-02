FROM golang:1.25.1-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/github.com/Bridgeless-Project/relayer-svc

ENV GO111MODULE="on"
ENV CGO_ENABLED=1
ENV GOOS="linux"
ENV GOPRIVATE=github.com/*
ENV GONOSUMDB=github.com/*
ENV GONOPROXY=github.com/*

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN go build  -o /usr/local/bin/relayer-svc /go/src/github.com/Bridgeless-Project/relayer-svc


FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/relayer-svc /usr/local/bin/relayer-svc

RUN apk add --no-cache ca-certificates

ENTRYPOINT ["relayer-svc"]