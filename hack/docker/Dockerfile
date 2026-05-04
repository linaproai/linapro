ARG BASE_IMAGE=alpine:3.22
FROM ${BASE_IMAGE}

###############################################################################
#                                INSTALLATION
###############################################################################

ENV WORKDIR=/app
ENV TZ=Asia/Shanghai
RUN apk add --no-cache ca-certificates tzdata 

COPY temp/image/lina $WORKDIR/lina
COPY apps/lina-core/manifest/config/config.template.yaml $WORKDIR/config.yaml

###############################################################################
#                                   START
###############################################################################
WORKDIR $WORKDIR
EXPOSE 8080
ENTRYPOINT ["./lina"]
