FROM docker.io/library/debian:oldstable-slim
RUN apt-get update
RUN apt-get install -y git python2.7
RUN apt-get install -y python-pip

RUN mkdir -p /opt/frickelclient
WORKDIR /opt/frickelclient
RUN git clone https://github.com/openlab-aux/lsmsd-frickelclient.git .
RUN pip install Flask==0.12.5 Flask-Script requests Flask-Wtf==0.8

ENTRYPOINT [ "python2", "manage.py", "runserver", "-h 0.0.0.0" ]