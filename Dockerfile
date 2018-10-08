FROM golang:1.11.0-alpine3.8 AS build
RUN apk add -U git
WORKDIR /go/src/github.com/Azure/kube-advisor
ADD Gopkg.toml Gopkg.lock ./
RUN go get -v github.com/golang/dep/cmd/dep && dep ensure -v -vendor-only
ADD kube_advisor.go .
RUN CGO_ENABLED=0 go install -a -ldflags '-w'

FROM alpine:3.8 AS run
# add GNU ps for -C, -o cmd, and --no-headers support
RUN apk --no-cache add procps
COPY --from=build /go/bin/kube-advisor /usr/local/bin/kube-advisor
CMD ["kube-advisor"]

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.build-date=$BUILD_DATE \
    org.label-schema.name="kube-advisor" \
    org.label-schema.description="Check if resource limits are applied to your containers" \
    org.label-schema.url="https://github.com/Azure/kube-advisor" \
    org.label-schema.vcs-ref=$VCS_REF \
    org.label-schema.vcs-url="https://github.com/Azure/kube-advisor" \
    org.label-schema.schema-version="1.0"
