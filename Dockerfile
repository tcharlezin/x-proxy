FROM alpine:latest

RUN mkdir /app

COPY x-proxy /app

CMD [ "/app/x-proxy" ]