FROM golang:1.23 as build

WORKDIR /build
ADD . /build

RUN go test ./...
RUN go build -o service ./cmd/service
RUN go build -o bot ./cmd/bot
RUN go build -o set-admin ./cmd/set-admin

# Use dedicated stage later
# FROM build as service
# COPY --from=build /build/app /srv/app

EXPOSE 8000
# WORKDIR /srv
