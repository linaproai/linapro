ARG BASE_IMAGE=alpine:3.22
FROM ${BASE_IMAGE}
ARG TARGETOS
ARG TARGETARCH

###############################################################################
#                                INSTALLATION
###############################################################################

ENV WORKDIR=/app

RUN apk add --no-cache ca-certificates tzdata

COPY temp/image/${TARGETOS}/${TARGETARCH}/lina $WORKDIR/lina
COPY .github/image/config.runtime.yaml $WORKDIR/config.yaml

###############################################################################
#                                   START
###############################################################################
WORKDIR $WORKDIR
EXPOSE 9120
CMD ["./lina"]
