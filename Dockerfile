FROM gliderlabs/alpine:3.2
MAINTAINER Ryan Eschinger <ryanesc@gmail.com>

COPY . /go/src/github.com/ryane/aws-keymaster

RUN apk add --update go git mercurial \
	&& cd /go/src/github.com/ryane/aws-keymaster \
	&& export GOPATH=/go \
	&& go get -t \
  && go test ./... \
	&& go build -o /bin/aws-keymaster \
	&& rm -rf /go \
	&& apk del --purge go git mercurial

ENTRYPOINT ["/bin/aws-keymaster"]
