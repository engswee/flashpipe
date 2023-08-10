FROM bitnami/minideb:bullseye
# ----------------------------------------
# Docker base image reference:
# https://hub.docker.com/r/bitnami/java
# ----------------------------------------
#https://github.com/bitnami/containers/blob/main/bitnami/java/1.8/debian-11/Dockerfile
#FROM docker.io/bitnami/minideb:bullseye
#https://hub.docker.com/r/bitnami/minideb

RUN install_packages ca-certificates git

ARG USER_HOME_DIR="/root"

COPY src/main/docker/gitconfig ${USER_HOME_DIR}/.gitconfig

COPY target/flashpipe /usr/bin/
#RUN chmod +x /usr/bin/flashpipe

#COPY src/main/docker/flashpipe.yaml ${USER_HOME_DIR}

