FROM golang:1.16 AS build
WORKDIR /pog
ADD . .
RUN echo $(pwd)
RUN CGO_ENABLED=0 GOOS=linux go build -o cmd/server ./cmd/server

FROM alpine:3.14 AS exec
EXPOSE 5555
COPY --from=build /pog/cmd/server/server .
CMD ["/server"]