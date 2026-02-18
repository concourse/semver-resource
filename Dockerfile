ARG base_image=cgr.dev/chainguard/wolfi-base
ARG builder_image=concourse/golang-builder

ARG BUILDPLATFORM
FROM --platform=${BUILDPLATFORM} ${builder_image} AS builder

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

COPY . /src
WORKDIR /src
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o /assets/in ./in
RUN go build -o /assets/out ./out
RUN go build -o /assets/check ./check
RUN set -e; for pkg in $(go list ./...); do \
    go test -o "/tests/$(basename $pkg).test" -c $pkg; \
    done

FROM ${base_image} AS resource
RUN apk --no-cache add \
    tzdata \
    ca-certificates \
    git \
    jq \
    openssh-client \
    cmd:ssh-keygen
RUN git config --global user.email "git@localhost"
RUN git config --global user.name "git"
COPY --from=builder assets/ /opt/resource/
RUN chmod +x /opt/resource/*

FROM resource AS tests
RUN apk --no-cache add bash
ARG SEMVER_TESTING_ACCESS_KEY_ID
ARG SEMVER_TESTING_SECRET_ACCESS_KEY
ARG SEMVER_TESTING_BUCKET
ARG SEMVER_TESTING_REGION
COPY --from=builder /tests /go-tests
WORKDIR /go-tests
RUN set -e; for test in /go-tests/*.test; do \
    $test; \
    done
COPY test/ /opt/resource-tests
RUN /opt/resource-tests/all.sh


FROM resource
