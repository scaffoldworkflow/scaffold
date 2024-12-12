FROM jfcarter2358/scaffold:build

# Copy in server source
COPY src/server /scaffold-build

# Build binary
RUN env GOOS=linux CGO_ENABLED=0 go test -c -cover -covermode=count -coverpkg=./... -o scaffold

# **************************************************************** #

FROM jfcarter2358/scaffold:run

# Copy built binary
COPY --from=0 /scaffold-build/scaffold ./

USER root

# Add start script and make it executable
ADD src/server/start-scaffold.sh /home/scaffold/start-scaffold.sh
RUN chmod +x /home/scaffold/start-scaffold.sh

# Copy over built UI files
COPY src/server/page/static /home/scaffold/static

# Own the entire scaffold home user
RUN chown -R scaffold:scaffold /home/scaffold

# Add scaffold to sudoers
RUN echo "scaffold     ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

USER scaffold
