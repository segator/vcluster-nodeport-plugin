FROM alpine
WORKDIR /
RUN mkdir -p /plugin
COPY bin/plugin /plugin/plugin