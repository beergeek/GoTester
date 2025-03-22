FROM debian:bookworm-slim

USER root

# Install dependencies
RUN wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz && rm -rf /usr/local/go && tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz

COPY weatherservice /usr/local/bin/weatherservice
EXPOSE 80

CMD [ "/usr/local/bin/weatherservice" ]