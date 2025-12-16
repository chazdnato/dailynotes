FROM alpine:3.23
ENTRYPOINT ["/usr/bin/dailynotes"]
COPY dailynotes /usr/bin
