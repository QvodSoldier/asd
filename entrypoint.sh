#!/bin/bash

mkdir -p /asd/data/agent
mv /asd/letmein /asd/data/agent/
mv /asd/asd_agent /asd/data/agent/
exec ./main
