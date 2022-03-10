FROM golang:1.12.17-alpine AS builder

RUN go env
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories &&\
    apk add --no-cache gcc musl-dev libpcap-dev &&\
    mkdir -p /go/src/github.com/zr-hebo/sniffer-agent
WORKDIR /go/src/github.com/zr-hebo/sniffer-agent
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o sniffer-agent .

# copy binary file
FROM alpine:3.15
RUN mkdir /app &&\
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories &&\
    apk add --no-cache libpcap-dev tzdata supervisor &&\
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime &&\
    sed -i "s/;nodaemon=false/nodaemon=true/" /etc/supervisord.conf

COPY --from=builder /go/src/github.com/zr-hebo/sniffer-agent/sniffer-agent /app/sniffer-agent
#COPY ./sniffer-agent.ini /etc/supervisor.d/sniffer-agent.ini
WORKDIR /app
ENTRYPOINT ["supervisord" ,"-c" ,"/etc/supervisord.conf"]