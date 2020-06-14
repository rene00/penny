FROM ubuntu:20.04

ENV FLASK_APP=penny

ENV LANG=en_AU.UTF-8

ENV PYTHONDONTWRITEBYTECODE 1

ENV PYTHONUNBUFFERED 1

WORKDIR /app

RUN apt update && DEBIAN_FRONTEND=noninteractive apt install -y python3 python \
    python3-pip gcc libffi-dev libxml2-dev libxslt1-dev redis-server \
    libssl-dev locales && apt-get clean all

RUN echo 'en_AU.UTF-8 UTF-8' > /etc/locale.gen && /usr/sbin/locale-gen

RUN mkdir -p /app/files/transactions /app/files/uploads

COPY . .

RUN pip3 install -r /app/requirements.txt && rm -rf /root/.cache/pip

ENTRYPOINT ["honcho", "--app-root=/app", "start"]

EXPOSE 5000
