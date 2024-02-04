FROM golang:1.21.6-alpine as build
RUN apk add --no-cache make cmake gcc g++ git openssl openssl-dev perl-test-harness-utils tcpdump libpcap libpcap-dev libbsd-dev perl-scope-guard perl-test-tcp curl bash
RUN curl -o /usr/bin/wait-for-it https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && chmod +x /usr/bin/wait-for-it
RUN mkdir -p /go/src/github.com/PROGNOSISTool/adapter-quic
ADD . /go/src/github.com/PROGNOSISTool/adapter-quic
WORKDIR /go/src/github.com/PROGNOSISTool/adapter-quic
ENV GOPATH /go
RUN go get
RUN cd $(ls -d /go/pkg/mod/github.com/!p!r!o!g!n!o!s!i!s!tool/pigotls*) && make
RUN cd $(ls -d /go/pkg/mod/github.com/!p!r!o!g!n!o!s!i!s!tool/ls-qpack-go*) && make
RUN go build -o /run_adapter bin/run_adapter/main.go

FROM alpine:3.14 as runtime
RUN apk add --no-cache jq tcpdump libpcap libpcap-dev
COPY --from=build /run_adapter /usr/bin/run_adapter
WORKDIR /root
ENTRYPOINT ["/usr/bin/run_adapter"]
