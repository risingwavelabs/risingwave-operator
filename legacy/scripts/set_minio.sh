#!/bin/bash
#
# Copyright 2022 Singularity Data
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -e

rm -f ${PREFIX_LOG}/mcli.log
touch ${PREFIX_LOG}/mcli.log

rm -f ${PREFIX_LOG}/minio.log
touch ${PREFIX_LOG}/minio.log

if [ ! -f "${PREFIX_BIN}/minio" ]; then
    echo "Binary minio not found in ${PREFIX_BIN}." >> ${PREFIX_LOG}/mcli.log
    exit 1
fi

if [ ! -f "${PREFIX_BIN}/mcli" ]; then
    echo "Binary mcli not found in ${PREFIX_BIN}." >> ${PREFIX_LOG}/mcli.log
    exit 1
fi

if [ ! ${MINIO_SERVER_PORT} ] || [ ! ${MINIO_CONSOLE_PORT} ]; then
	  echo "Environment variable MINIO_SERVER_PORT, MINIO_CONSOLE_PORT does not exists." >> ${PREFIX_LOG}/mcli.log
	  exit 1
fi

nohup ${PREFIX_BIN}/minio server ${PREFIX_DATA}/hummock --address 0.0.0.0:${MINIO_SERVER_PORT} --console-address 0.0.0.0:${MINIO_CONSOLE_PORT} --config-dir ${PREFIX_CONFIG}/minio > ${PREFIX_LOG}/minio.log &

echo "Minio Server is now running." >> ${PREFIX_LOG}/mcli.log
while :
do
    if [ $(curl --write-out %{http_code} --silent --output /dev/null http://127.0.0.1:${MINIO_SERVER_PORT}/minio/health/live) -ne 200 ]; then
        echo "Waiting for minio server to be ready."
        sleep 1
    else
        break
    fi
done

echo "Minio Server is ready." >> ${PREFIX_LOG}/mcli.log

{
    ${PREFIX_BIN}/mcli -C ${PREFIX_CONFIG}/mcli alias set hummock-minio http://127.0.0.1:${MINIO_SERVER_PORT} hummockadmin hummockadmin >> ${PREFIX_LOG}/mcli.log 2>&1
    ${PREFIX_BIN}/mcli -C ${PREFIX_CONFIG}/mcli admin user add hummock-minio/ hummock 12345678 >> ${PREFIX_LOG}/mcli.log 2>&1
    ${PREFIX_BIN}/mcli -C ${PREFIX_CONFIG}/mcli admin policy set hummock-minio/ readwrite user=hummock >> ${PREFIX_LOG}/mcli.log 2>&1
    ${PREFIX_BIN}/mcli -C ${PREFIX_CONFIG}/mcli mb hummock-minio/hummock001
} >> ${PREFIX_LOG}/mcli.log 2>&1

echo "Minio Server has been Set." >> ${PREFIX_LOG}/mcli.log
echo "Minio Server ready."

touch ${PREFIX_LOG}/minio_server_ready

while :
do
    sleep 1
done