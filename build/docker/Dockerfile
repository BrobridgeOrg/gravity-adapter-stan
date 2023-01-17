FROM golang:1.17.13-alpine3.16 AS builder
WORKDIR /
COPY . .

RUN go build -o /gravity-adapter-stan ./cmd/gravity-adapter-stan/gravity-adapter-stan.go

FROM alpine:3.17.1
WORKDIR /
RUN apk update && apk upgrade --available
COPY --from=builder /gravity-adapter-stan /gravity-adapter-stan
COPY ./configs /configs
COPY ./settings /settings
COPY ./build/docker/startup.sh /startup.sh

RUN chmod 700 -R /settings && chown -R 1001  /settings
USER 1001

CMD ["/bin/sh", "/startup.sh"]
