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
COPY .github/workflows/image-build/config.runtime.yaml $WORKDIR/config.yaml



# init and mock commands are used to prepare the runtime environment, 
# such as creating necessary directories and files. 
# This ensures that when the container starts, it has all the required setup to run properly.
RUN $WORKDIR/lina init --confirm=init
RUN $WORKDIR/lina mock --confirm=mock

###############################################################################
#                                   START
###############################################################################
WORKDIR $WORKDIR
EXPOSE 8080
CMD ["./lina"]
