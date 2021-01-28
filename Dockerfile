FROM golang:1.15-alpine AS dev

RUN apk add --no-cache \
    g++ \
    inotify-tools

WORKDIR /go/src/github.com/kaizendorks/terraform-cloud-exporter

ENTRYPOINT ["./hot-reload.sh"]

FROM golang:1.15-alpine AS build

WORKDIR /go/src/github.com/kaizendorks/terraform-cloud-exporter

COPY . .
RUN go build

FROM alpine:3 AS prod

COPY --from=build /go/src/github.com/kaizendorks/terraform-cloud-exporter/terraform-cloud-exporter /bin/terraform-cloud-exporter

USER nobody
ENTRYPOINT [ "/bin/terraform-cloud-exporter" ]
