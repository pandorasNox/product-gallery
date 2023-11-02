FROM migrate/migrate:v4.16.2

COPY migrations /migrations

RUN mkdir /scripts
COPY cicd/container-images/migrate/migrate.sh /scripts/migrate.sh
