FROM golang:latest as BUILD
WORKDIR /go/src/github.com/mwmahlberg/mgogenerate
COPY . /go/src/github.com/mwmahlberg/mgogenerate
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mgogenerate .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /usr/local/bin
COPY --from=BUILD /go/src/github.com/mwmahlberg/mgogenerate/mgogenerate .
CMD ["/usr/local/bin/mgogenerate"]