FROM --platform=$BUILDPLATFORM golang:1.21 as builder
WORKDIR $GOPATH/src/github.com/davidepedranz/go-hole/
COPY . ./
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /go-hole

FROM gcr.io/distroless/static-debian12
WORKDIR /app/
COPY --from=builder /go-hole ./go-hole
COPY config.json.default /etc/gohole/config.json
COPY static ./static
ENTRYPOINT ["./go-hole"]
EXPOSE 53/udp 8080
