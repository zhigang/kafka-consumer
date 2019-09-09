FROM golang:1.12 AS builder
WORKDIR /go/src/zprojects/kafka-consumer
COPY ./ /go/src/zprojects/kafka-consumer/
# disable cgo 
ENV CGO_ENABLED=0
# build steps
RUN echo ">>> 1: go version" && go version
RUN echo ">>> 2: go get" && go get -v -d
RUN echo ">>> 3: go install" && go install

# make application docker image use alpine
FROM  alpine:3.10
RUN apk --no-cache add ca-certificates
WORKDIR /go/bin/
# copy config file to image (like config.json or config.staging.json)
RUN mkdir config
COPY --from=builder /go/src/zprojects/kafka-consumer/config/config.yml ./config/
# copy execute file to image
COPY --from=builder /go/bin/ ./
EXPOSE 3000
CMD ["./kafka-consumer"]
