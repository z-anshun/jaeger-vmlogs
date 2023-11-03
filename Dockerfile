FROM docker.io/library/alpine:3.16

ADD jaeger-vl-linux-amd64 /go/bin/jaeger-vl

RUN mkdir /plugin

# /plugin/ location is defined in jaeger-operator
CMD ["cp", "/go/bin/jaeger-vl", "/plugin/jaeger-vl"]
