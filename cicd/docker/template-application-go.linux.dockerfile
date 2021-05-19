ARG ARCH=amd64

FROM arhatdev/builder-go:alpine as builder
FROM arhatdev/go:alpine-${ARCH}
ARG APP=meeting-minutes-bot

ENTRYPOINT [ "/meeting-minutes-bot" ]
