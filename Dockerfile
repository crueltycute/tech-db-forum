FROM ubuntu:18.04

MAINTAINER nmerk

ENV PGVER 10
RUN apt-get -y update && apt-get install -y postgresql-$PGVER

RUN apt install -y golang-1.10 git
ENV GOROOT /usr/lib/go-1.10
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR $GOPATH/src/github.com/crueltycute/tech-db-forum
ADD . $GOPATH/src/github.com/crueltycute/tech-db-forum

USER postgres
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql docker -f init.sql &&\
    /etc/init.d/postgresql stop


USER root

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf
EXPOSE 5432
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

#RUN go get ./...
#CMD service postgresql start && go run ./cmd/forum-server/main.go --scheme=http --port=5000 --host=0.0.0.0
RUN ls
RUN go build cmd/main.go
CMD service postgresql start && ./main