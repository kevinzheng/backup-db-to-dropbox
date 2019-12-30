FROM alpine:latest

RUN apk --update add postgresql-client

WORKDIR /app

COPY target/backup-db-to-dropbox /app/backup-db-to-dropbox

# copy crontabs for root user
COPY cronjobs /etc/crontabs/root

# start crond with log level 8 in foreground, output to stderr
CMD ["crond", "-f", "-d", "8"]