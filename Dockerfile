FROM golang as builder

RUN useradd -r appuser

COPY src $GOPATH/src/depends-on/
WORKDIR $GOPATH/src/depends-on/

RUN go get -d -v

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/depends-on

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/depends-on /go/bin/depends-on

USER appuser

ENTRYPOINT ["/go/bin/depends-on"]
