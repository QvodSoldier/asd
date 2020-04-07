FROM index.alauda.cn/alaudaorg/centos:7

USER root

RUN mkdir -p /asd/data

COPY ./main /asd/

COPY ./letmein /asd/

COPY ./asd_agent /asd/

ADD ./entrypoint.sh /asd/

RUN chmod 777 /asd/entrypoint.sh

WORKDIR /asd

EXPOSE 12580

ENTRYPOINT ["./entrypoint.sh"]
