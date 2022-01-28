FROM golang:1.17
ADD . /go/src/github.com/geezyx/sudo-server/
WORKDIR /go/src/github.com/geezyx/sudo-server/
RUN CGO_ENABLED=0 GOOS=linux go build -o sudo-server ./cmd/server

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /go/src/github.com/geezyx/sudo-server/sudo-server ./
CMD ["./sudo-server"]