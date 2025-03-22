FROM debian:bookworm-slim

USER root

# Install dependencies
RUN apt update && apt install -y wget
RUN wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
RUN rm -rf /usr/local/go
RUN tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz

COPY weatherservice /usr/local/bin/weatherservice
EXPOSE 80

CMD [ "/usr/local/bin/weatherservice" ]