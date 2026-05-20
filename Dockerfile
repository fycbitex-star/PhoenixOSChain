FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/phoenixd ./cmd/phoenixd

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/phoenixd /usr/local/bin/phoenixd
COPY configs ./configs
ENTRYPOINT ["phoenixd"]
