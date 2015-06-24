# Pull base image.
FROM google/golang

# Install HG for go get
#RUN apt-get update && \
#    apt-get install -y mercurial curl git


ADD . /gopath/src/github.com/wayt/imgbot
WORKDIR /gopath/src/github.com/wayt/imgbot

RUN go get
RUN go install

ENTRYPOINT ["/gopath/bin/imgbot"]

EXPOSE 8080
