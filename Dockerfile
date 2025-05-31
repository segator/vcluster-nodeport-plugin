FROM alpine:3.18

COPY bin/plugin /plugin

ENTRYPOINT ["/plugin"]