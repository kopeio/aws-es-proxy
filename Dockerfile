FROM debian:jessie

RUN apt-get update && apt-get install --yes ca-certificates

COPY .build/aws-es-proxy /usr/bin/aws-es-proxy

CMD /usr/bin/aws-es-proxy

