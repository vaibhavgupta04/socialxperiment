FROM postgres:latest

ENV POSTGRES_USER="admin"
ENV POSTGRES_PASSWORD="admin"
ENV POSTGRES_DB="socex"
# Expose the default  Postgres port
EXPOSE 5432 5432

# Optionally, you can add an initialization script to create tables or seed data
# COPY ./init.sql /docker-entrypoint-initdb.d/
