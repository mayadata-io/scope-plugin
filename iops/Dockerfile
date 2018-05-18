FROM ubuntu 
ENV CORTEXAGENT=""
RUN apt-get -y update   
RUN apt-get install -y curl
COPY ./iops /usr/bin/iops
LABEL works.weave.role=system
ENTRYPOINT ["/usr/bin/iops"]


