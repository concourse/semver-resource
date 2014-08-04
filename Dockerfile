FROM ubuntu:14.04

RUN apt-get update && apt-get -y install wget

ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /tmp/go

ADD . /tmp/go/src/github.com/concourse/semver-resource

ENV GOPATH /tmp/go/src/github.com/concourse/semver-resource/Godeps/_workspace:$GOPATH

RUN wget -qO- https://storage.googleapis.com/golang/go1.3.linux-amd64.tar.gz | tar -C /usr/local -xzf - && \
      go build -o /opt/resource/check github.com/concourse/semver-resource/check && \
      go build -o /opt/resource/in github.com/concourse/semver-resource/in && \
      go build -o /opt/resource/out github.com/concourse/semver-resource/out && \
      rm -rf /tmp/go /usr/local/go
