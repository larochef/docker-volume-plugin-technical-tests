FROM ubuntu:14.04

COPY docker-volume-plugin-technical-tests /usr/bin/docker-volume-plugin-technical-tests
RUN chmod +x /usr/bin/docker-volume-plugin-technical-tests

VOLUME /data
VOLUME /run
VOLUME /dev

ENTRYPOINT ["/usr/bin/docker-volume-plugin-technical-tests", "-root", "/data", "-path", "/tmp"]
