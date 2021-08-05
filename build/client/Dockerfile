FROM golang:1.16 AS build
WORKDIR /pog
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cmd/client ./cmd/client

FROM alpine:3.14 AS exec
COPY --from=build /pog/cmd/client/client .
CMD ["/client"]