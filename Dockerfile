FROM alpine:3.22
ENTRYPOINT ["/usr/bin/dailynotes"]
COPY dailynotes /usr/bin
