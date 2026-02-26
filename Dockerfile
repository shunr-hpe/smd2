# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

# FROM docker.io/library/alpine:3.15
# FROM docker.io/library/alpine:3.23
FROM rockylinux:9.3.20231119

ENV SMD_PORT=8080

# add user on alpine linux
# RUN addgroup -g 1000 smd && \
#     adduser -D -u 1000 -G smd smd && \
#     chown -R smd:smd /home/smd

# add user on rocky
RUN useradd -m smd

WORKDIR /home/smd

COPY bin/smd2-server /usr/local/bin/smd2-server

USER smd

RUN mkdir -p data

ENTRYPOINT ["sh", "-c", "/usr/local/bin/smd2-server serve --port $SMD_PORT --database-url file:data/smd2.db?_fk=1"]
