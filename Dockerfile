# 编译website
FROM golang:1.13.0-alpine3.10 AS builder

ADD . /go/src/github.com/eudore/godoc
RUN apk add git && \
	go version && go env && \
	cd /go/src/github.com/eudore/godoc && \
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" app.go && \
	git clone https://github.com/eudore/godoc.wiki.git && \
	cp -rf /usr/local/go .

# 创建运行镜像
FROM alpine:latest

RUN apk add git

COPY --from=builder /go/src/github.com/eudore/godoc/ /

CMD ["/app", "--goroot=/go", "--data=/data", "--data=/godoc.wiki"]
