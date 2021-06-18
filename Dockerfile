FROM alpine:latest

RUN apk --update add postgresql-client
RUN apk --update add mysql-client

WORKDIR /app

COPY target/linux/backupdbtodropbox /app/backupdbtodropbox

CMD ["/app/backupdbtodropbox", "-k"]