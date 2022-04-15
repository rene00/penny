FROM ubuntu:20.04

ENV FLASK_APP=penny

ENV LANG=en_AU.UTF-8

ENV PYTHONDONTWRITEBYTECODE 1

ENV PYTHONUNBUFFERED 1

ENV PATH /root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

WORKDIR /app

RUN apt update && DEBIAN_FRONTEND=noninteractive apt install -y \
    python3.9-venv python3.9-minimal python3.9-dev gcc libffi-dev libxml2-dev \
    libxslt1-dev redis-server libssl-dev locales curl && \
    apt-get clean all && \
    curl -sSL https://install.python-poetry.org | /usr/bin/python3.9 - --version 1.1.13 

RUN echo 'en_AU.UTF-8 UTF-8' > /etc/locale.gen && /usr/sbin/locale-gen

RUN mkdir -p /app/files/transactions /app/files/uploads

COPY . .

RUN poetry install --no-dev

ENTRYPOINT ["poetry", "run", "honcho", "--app-root=/app", "start"]

EXPOSE 5000
