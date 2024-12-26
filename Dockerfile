FROM golang:1.23 as build

WORKDIR /build
ADD . /build

RUN go test ./...
RUN go build -o service ./cmd/service

FROM build as service

COPY --from=build /build/app /srv/app

EXPOSE 8080
WORKDIR /srv

CMD ["/srv/service"]
