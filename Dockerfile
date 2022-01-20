FROM alpine:3.15 AS build
ARG VERSION=1.16.1

RUN \
  mkdir -p /pgbouncer && \
  # add build dependencies
  apk --update --no-cache add \
    build-base \
    curl \
    c-ares-dev \
    libevent-dev \
    openssl-dev \
    pkgconfig && \
  # Download
  curl -o /tmp/pgbouncer-$VERSION.tar.gz \
    -L https://pgbouncer.github.io/downloads/files/$VERSION/pgbouncer-$VERSION.tar.gz && \
  # Unpack
  cd /tmp && \
  tar xvfz /tmp/pgbouncer-$VERSION.tar.gz && \
  # Compile
  cd pgbouncer-$VERSION && \
  ./configure --prefix=/pgbouncer && \
  make && \
   # Move binary
  cp pgbouncer /pgbouncer

FROM alpine:3.15

# install pgbouncer dependencies
RUN apk --update --no-cache add c-ares libevent openssl

# create application directory, user, group, and change ownership
RUN addgroup postgres && \
  adduser -D --ingroup postgres postgres && \
  mkdir -p /pgbouncer && \
  chown -R postgres:postgres /pgbouncer/

# copy pgbouncer and consul-template from build stage
COPY --from=build --chown=postgres:postgres /pgbouncer/pgbouncer /pgbouncer/

WORKDIR /pgbouncer
USER postgres
EXPOSE 5432
ENTRYPOINT ["/pgbouncer/pgbouncer"]
CMD ["/etc/pgbouncer/pgbouncer.ini"]
