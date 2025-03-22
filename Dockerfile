FROM debian:bookworm-slim

USER root

COPY weatherservice /usr/local/bin/weatherservice
EXPOSE 8080

CMD [ "/usr/local/bin/weatherservice" ]