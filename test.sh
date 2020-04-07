#!/bin/bash

cd /home/mahuang/workspace/gopath/src/ggstudy/asd/agent

go build asd_agent.go

mv asd_agent ../

cd letmein

gcc --static -o letmein letmein.c

mv letmein ../../

cd ../../

go build main.go
