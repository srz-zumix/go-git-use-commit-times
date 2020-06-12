
FROM ubuntu:20.04
ARG LIBGIT2_TAG=v1.0.1

LABEL maintainer "srz_zumix <https://github.com/srz-zumix>"

ENV DEBIAN_FRONTEND=noninteractive
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update -q -y && \
    apt-get install -y --no-install-recommends software-properties-common apt-transport-https && \
    apt-get update -q -y && \
    apt-get install -y \
        golang \
        wget curl libssl-dev pkg-config time \
        git make cmake ca-certificates build-essential && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /usr/local/src
RUN echo ${LIBGIT2_TAG} && \
    git clone https://github.com/libgit2/libgit2.git -b ${LIBGIT2_TAG} && \
    git clone https://github.com/srz-zumix/git-use-commit-times.git
WORKDIR /usr/local/src/libgit2/build
RUN cmake .. -DCMAKE_INSTALL_PREFIX=/usr && cmake --build . --target install

WORKDIR /usr/local/src/git-use-commit-times
RUN go install
