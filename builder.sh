#!/bin/sh

make out/executor
docker build -t gcr.io/priya-wadhwa/executor .
docker push gcr.io/priya-wadhwa/executor