FROM python:3.12.8-slim-bookworm

WORKDIR /home/scaffold

USER root

RUN apt-get update && apt-get install -y \
    jq

RUN adduser --disabled-password scaffold

SHELL ["/bin/bash", "-c"]

# Add scaffold user to sudoers
RUN echo "scaffold     ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

RUN mkdir -p /home/scaffold/.ssh

COPY src/worker/test/foo.sh /tmp/foo.sh
RUN chmod +x /tmp/foo.sh

# Own the entire scaffold home user
RUN chown -R scaffold:scaffold /home/scaffold

USER scaffold
