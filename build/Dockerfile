FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN apk update && \
    apk upgrade && \
    apk add --no-cache bash git openssh

RUN mkdir /etc/sofi
WORKDIR /etc/sofi

ADD sofi /etc/sofi

RUN chmod 655 /etc/sofi/sofi

ENTRYPOINT ["/etc/sofi/sofi"]
EXPOSE 8080