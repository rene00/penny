FROM ubuntu:16.04

ARG http_proxy="http://10.0.1.13:3128"

ARG https_proxy="http://10.0.1.13:3128"

ARG pip_trusted_host="hound.compounddata.com"

ENV pip_index_url "http://hound.compounddata.com:6543/simple"

ENV FLASK_APP=penny.py

ENV CONFIG_FILE=conf.py

ENV LANG=en_US.UTF-8

WORKDIR /app

RUN echo 'Acquire::http::Proxy "http://10.0.1.13:3128";' > /etc/apt/apt.conf

RUN http_proxy=$http_proxy apt update && apt install -y python3 python \
    python3-pip gcc libffi-dev libxml2-dev libxslt1-dev redis-server \
    libssl-dev locales

RUN echo 'en_US.UTF-8 UTF-8' > /etc/locale.gen && /usr/sbin/locale-gen

RUN mkdir -p /app/files/transactions /app/files/uploads

COPY . .

RUN pip3 install -r /app/requirements.txt --index-url ${pip_index_url} \
    --trusted-host ${pip_trusted_host} honcho

CMD honcho --app-root=/app start

EXPOSE 5000
