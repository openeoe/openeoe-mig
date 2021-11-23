#!/bin/bash
#docker_registry_ip="10.0.3.40:5005"
docker_registry_ip="index.docker.io"
docker_id="openeoe"
controller_name="openeoe-migration"
controller_version="v0.0.2"
#controller_version="v2.0.2.c"

export GO111MODULE=on
go mod vendor

go build -o build/_output/bin/$controller_name -gcflags all=-trimpath=`pwd` -asmflags all=-trimpath=`pwd` -mod=vendor openeoe/openeoe/openeoe-migration/cmd/manager && \

docker build -t $docker_registry_ip/$docker_id/$controller_name:$controller_version build && \
docker push $docker_registry_ip/$docker_id/$controller_name:$controller_version
