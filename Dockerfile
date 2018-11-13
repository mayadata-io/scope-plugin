FROM ubuntu 

RUN apt-get -y update

RUN apt-get install -y curl

COPY ./scope-plugin /usr/bin/scope-plugin

ENTRYPOINT ["/usr/bin/scope-plugin"]
