# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

# FROM docker.io/library/alpine:3.15
# FROM docker.io/library/alpine:3.23
FROM rockylinux:9.3.20231119

# add user on alpine linux
# RUN addgroup -g 1000 smd && \
#     adduser -D -u 1000 -G smd smd && \
#     chown -R smd:smd /home/smd

# add user on rocky
RUN useradd -m smd

WORKDIR /home/smd

COPY bin/smd2 /usr/local/bin/smd2

USER smd

ENTRYPOINT ["/usr/local/bin/smd2"]
