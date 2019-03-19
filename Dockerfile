FROM golang:1.11.5 AS build

ENV DEP_VERSION=v0.5.0

RUN curl https://raw.githubusercontent.com/golang/dep/${DEP_VERSION}/install.sh | sh

ENV WORKDIR=/go/src/github.com/3scale/3scale-operator
ADD . ${WORKDIR}
WORKDIR ${WORKDIR}

RUN dep ensure -v

ENV OOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

RUN go build -o build/bin/3scale-operator cmd/manager/main.go

FROM alpine:3.8

ENV OPERATOR=/usr/local/bin/3scale-operator \
    USER_UID=1001 \
    USER_NAME=3scale-operator

# install operator binary
COPY --from=build /go/src/github.com/3scale/3scale-operator/build/bin/_output/3scale-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

