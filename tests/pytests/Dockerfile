# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

FROM python:3.14.2-alpine3.23

RUN apk update && \
    apk add --no-cache \
        python3-dev \
        pytest \
        curl \
        jq \
        bash

COPY requirements.txt .
RUN pip install --upgrade pip
RUN pip install -r requirements.txt

WORKDIR /app

COPY . .

CMD [ "/usr/local/bin/pytest", "/app" ]
