FROM debian:bookworm-slim

USER root

COPY weatherservice /usr/local/bin/weatherservice
EXPOSE 80

CMD [ "/usr/local/bin/weatherservice" ]