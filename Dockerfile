FROM golang:latest as BUILD
WORKDIR /go/src/github.com/mwmahlberg/mongen
COPY . /go/src/github.com/mwmahlberg/mongen
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mongen .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /usr/local/bin
COPY --from=BUILD /go/src/github.com/mwmahlberg/mongen/mongen .
CMD ["/usr/local/bin/mongen"]