FROM alpine:3.4

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY /.build/artifacts/aws-es-proxy /usr/bin/aws-es-proxy

CMD /usr/bin/aws-es-proxy

