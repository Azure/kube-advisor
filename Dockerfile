FROM golang:1.9 AS build
WORKDIR /go/src/github.com/Azure/kube-resource-checker
ADD Gopkg.toml Gopkg.lock ./
RUN go get -v github.com/golang/dep/cmd/dep && dep ensure -v -vendor-only
ADD main.go .
RUN CGO_ENABLED=0 go install -a -ldflags '-w'

FROM alpine:3.7 AS run
# add GNU ps for -C, -o cmd, and --no-headers support
RUN apk --no-cache add procps
COPY --from=build /go/bin/kube-resource-checker /usr/local/bin/kube-resource-checker
CMD ["kube-resource-checker"]

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.build-date=$BUILD_DATE \
    org.label-schema.name="kube-resource-checker" \
    org.label-schema.description="Check if resource limits are applied to your containers" \
    org.label-schema.url="https://github.com/Azure/kube-resources-checker" \
    org.label-schema.vcs-ref=$VCS_REF \
    org.label-schema.vcs-url="https://github.com/Azure/kube-resources-checker" \
    org.label-schema.schema-version="1.0"
