FROM golang:1.18.1-alpine

RUN apk update && apk add git

WORKDIR /scaffold-build
COPY src/server /scaffold-build
RUN env GOOS=linux CGO_ENABLED=0 go build -v -o scaffold

# FROM alpine:latest
# FROM ubuntu:latest
FROM ubuntu:22.04



WORKDIR /home/scaffold

COPY --from=0 /scaffold-build/scaffold ./

ENV DEBIAN_FRONTEND=noninteractive

# RUN apk update \
#     && apk add \
RUN apt-get update \
    && apt-get install -y \
    bash \
    git \
    curl \
    podman \
    # openssh \
    # docker \
    fuse-overlayfs \
    adduser \
    ca-certificates \
    gnupg \
    lsb-release \
    tzdata \
    software-properties-common
    # openrc

# Install Docker
# ADD helpers/install-docker.sh /tmp/install-docker.sh
# RUN /tmp/install-docker.sh

RUN adduser --disabled-password scaffold

# RUN echo http://dl-2.alpinelinux.org/alpine/edge/community/ >> /etc/apk/repositories \
#     && apk add -U shadow

SHELL ["/bin/bash", "-c"]

RUN mkdir -p /home/scaffold/.ssh

ADD src/server/start-scaffold.sh /home/scaffold/start-scaffold.sh
RUN chmod u+x /home/scaffold/start-scaffold.sh

COPY src/server/ui-dist /home/scaffold

RUN chown -R scaffold:scaffold /home/scaffold

RUN usermod --add-subuids 200000-201000 --add-subgids 200000-201000 scaffold
# RUN rc-update add cgroups

# RUN echo "tun >>/etc/modules"
# RUN echo "scaffold:100000:65536 >/etc/subuid"
# RUN echo "scaffold:100000:65536 >/etc/subgid"

# Add script to startup Docker
# COPY helpers/start-docker.sh /bin/start-docker.sh
# RUN chmod +x /bin/start-docker.sh

# USER root
USER scaffold

CMD ["./start-scaffold.sh"]
