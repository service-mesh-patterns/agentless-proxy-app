FROM golang:1.16.3 as build

RUN mkdir /go/app

COPY ./ /go/app

WORKDIR /go/app

RUN CGO_ENABLED=0 go build -o ./bin/server .

FROM alpine:latest

RUN mkdir /app
COPY --from=build /go/app/bin/server /app/server

ENTRYPOINT ["/app/server"]