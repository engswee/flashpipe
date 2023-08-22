FROM bitnami/minideb:bullseye
# ----------------------------------------
# Docker base image reference:
# https://hub.docker.com/r/bitnami/minideb
# ----------------------------------------

RUN install_packages ca-certificates git

ARG USER_HOME_DIR="/root"

COPY build/gitconfig ${USER_HOME_DIR}/.gitconfig

COPY output/flashpipe /usr/bin/flashpipe
