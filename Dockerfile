FROM ubuntu:18.04

MAINTAINER nmerk

ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get -y update
RUN apt install -y git wget gcc gnupg

ENV PGVER 11

RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ bionic-pgdg main" > /etc/apt/sources.list.d/pgdg.list

# get the signing key and import it
RUN wget https://www.postgresql.org/media/keys/ACCC4CF8.asc
RUN apt-key add ACCC4CF8.asc

# fetch the metadata from the new repo
RUN apt-get update

RUN apt-get install -y  postgresql-$PGVER

RUN wget https://dl.google.com/go/go1.12.linux-amd64.tar.gz
RUN tar -xvf go1.12.linux-amd64.tar.gz
RUN mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH $HOME/go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /server
COPY . .

USER postgres
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql docker -f init.sql &&\
    /etc/init.d/postgresql stop

USER root

FROM ubuntu:18.04

MAINTAINER nmerk

ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get -y update
RUN apt install -y git wget gcc gnupg

ENV PGVER 11

RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ bionic-pgdg main" > /etc/apt/sources.list.d/pgdg.list

# get the signing key and import it
RUN wget https://www.postgresql.org/media/keys/ACCC4CF8.asc
RUN apt-key add ACCC4CF8.asc

# fetch the metadata from the new repo
RUN apt-get update

RUN apt-get install -y  postgresql-$PGVER

RUN wget https://dl.google.com/go/go1.12.linux-amd64.tar.gz
RUN tar -xvf go1.12.linux-amd64.tar.gz
RUN mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH $HOME/go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /server
COPY . .

USER postgres
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql docker -f init.sql &&\
    /etc/init.d/postgresql stop

USER root

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "synchronous_commit = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "fsync = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "random_page_cost = 1.0" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "work_mem = 10MB" >> /etc/postgresql/$PGVER/main/postgresql.conf

EXPOSE 5432
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

RUN go build cmd/main.go
CMD service postgresql start && ./main

EXPOSE 5432
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

RUN go build cmd/main.go
CMD service postgresql start && ./main