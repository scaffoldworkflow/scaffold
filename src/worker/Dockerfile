FROM python:3.12.8-slim-bookworm

WORKDIR /home/scaffold

USER root

RUN apt-get update && apt-get install -y \
    jq

RUN adduser --disabled-password scaffold

SHELL ["/bin/bash", "-c"]

# Add start script and make it executable
ADD src/worker/.executor.sh /home/scaffold/.executor.sh
RUN chmod +x /home/scaffold/.executor.sh

# Add our headers
ADD src/worker/.header.py /home/scaffold/.header.py
ADD src/worker/.header.sh /home/scaffold/.header.sh

# Add scaffold user to sudoers
RUN echo "scaffold     ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

RUN mkdir -p /home/scaffold/.ssh

# Own the entire scaffold home user
RUN chown -R scaffold:scaffold /home/scaffold

USER scaffold

CMD ["./.executor.sh"]
