#!/usr/bin/env bash

set -ex

# Docker requires root on my system. I awkwardly have to propagate config to the
# root user, but use kubectl configured as my normal user.
sudo KO_DOCKER_REPO=docker.io/spencerjp $(which ko) resolve -f conf/surfdash-deployment.yaml --platform=linux/arm64 | kubectl apply -f -

kubectl apply -f conf/surfdash-ingress.yaml -f conf/surfdash-service.yaml
