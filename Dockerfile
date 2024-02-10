FROM golang:1.21 as builder
WORKDIR $GOPATH/src/github.com/davidepedranz/go-hole/
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o /go-hole

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go-hole /go-hole
COPY /data/test-blacklist.txt /data/blacklist.txt
ENTRYPOINT ["/go-hole"]
EXPOSE 53/udp
