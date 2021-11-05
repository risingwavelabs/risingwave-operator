FROM ubuntu:20.04 as builder

ENV LANG en_US.utf8

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y install curl tmux

ARG TARGETARCH

ENV MINIO_SYSTEM="linux-${TARGETARCH}"

ENV PREFIX="/risingwave"
ENV PREFIX_BIN="${PREFIX}/bin" \
    PREFIX_CONFIG="${PREFIX}/config" \
    PREFIX_DATA="${PREFIX}/data" \
    PREFIX_LOG="${PREFIX}/log"

RUN mkdir -p ${PREFIX_BIN} ${PREFIX_CONFIG} ${PREFIX_DATA} ${PREFIX_LOG} ${PREFIX_CONFIG}/mcli

ENV MCLI_DOWNLOAD_BIN="https://dl.min.io/client/mc/release/${MINIO_SYSTEM}/mc" \
    MINIO_DOWNLOAD_BIN="https://dl.min.io/server/minio/release/${MINIO_SYSTEM}/minio"

RUN curl -fL -o ${PREFIX_BIN}/minio ${MINIO_DOWNLOAD_BIN}
RUN chmod +x ${PREFIX_BIN}/minio

RUN curl -fL -o ${PREFIX_BIN}/mcli ${MCLI_DOWNLOAD_BIN}
RUN chmod +x ${PREFIX_BIN}/mcli

ENV MINIO_PROMETHEUS_AUTH_TYPE="public" \
    MINIO_ROOT_PASSWORD="hummockadmin" \
    MINIO_ROOT_USER="hummockadmin"
# default ports
ENV MINIO_SERVER_PORT="9301" \
    MINIO_CONSOLE_PORT="9400"

RUN touch ${PREFIX_LOG}/mcli.log ${PREFIX_LOG}/minio.log

COPY scripts/set_minio.sh ${PREFIX_BIN}/set_minio.sh

RUN chmod 777 ${PREFIX_BIN}/set_minio.sh

EXPOSE 9301 9400

ENTRYPOINT [ "/bin/bash", "-c" ]
CMD [ "${PREFIX_BIN}/set_minio.sh" ]


