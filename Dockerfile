# API build stage
FROM golang:1.18.3-alpine3.16 as go-builder
ARG GOPROXY=goproxy.cn

ENV GOPROXY=https://${GOPROXY},direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /data

COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN go build -o ./bin/eapi cmd/eapi/eapi.go

# Fianl running stage
FROM golang:1.18.3-alpine3.16
LABEL maintainer="goproxy@gotomicro.com"

WORKDIR /data

COPY --from=go-builder /data/bin/eapi /bin/
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

CMD ["sh", "-c", "/bin/eapi"]
