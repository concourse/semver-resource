FROM golang:1.8.3-alpine

ENV CONCOURSE_CODE_PATH ${GOPATH}/src/github.com/concourse/semver-resource

RUN apk add --update git bash \
  && rm -rf /var/cache/apk/*

ADD . /code

RUN mkdir -p $(dirname ${CONCOURSE_CODE_PATH}) \
    && ln -s /code ${CONCOURSE_CODE_PATH}

RUN cd ${CONCOURSE_CODE_PATH} \
  && go get -v -d ./...

RUN cd ${CONCOURSE_CODE_PATH} \
  && ./scripts/build

RUN cd ${CONCOURSE_CODE_PATH} \
  && mkdir -p /opt/resource \
  && cp assets/* /opt/resource

RUN rm -rf ${GOPATH} ${GOROOT} /usr/local/go /code
