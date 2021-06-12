FROM centos 

COPY server /server 

COPY run.sh  /run.sh

#WORKDIR /

RUN yum install -y curl

EXPOSE 50051

RUN chmod +x run.sh

#CMD [/bin/sh -x /run.sh]
#CMD [/bin/sh -C run.sh ]

ENTRYPOINT ["/run.sh"]

#CMD [/opt/mongodb_exporter --mongodb.uri=mongodb://admin:mep7Z9wv5X@10.0.2.102:27017]
