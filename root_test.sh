#!/bin/bash
docker rmi index.alauda.cn/alaudaorg/asd:v2

docker build -f Dockerfile -t index.alauda.cn/alaudaorg/asd:v2 .

docker push index.alauda.cn/alaudaorg/asd:v2
