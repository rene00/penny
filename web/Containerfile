FROM ubuntu:20.04

ENV FLASK_APP=penny

ENV LANG=en_AU.UTF-8

ENV PYTHONDONTWRITEBYTECODE 1

ENV PYTHONUNBUFFERED 1

ENV PYENV_ROOT /root/.pyenv

ENV PATH $PYENV_ROOT/shims:$PYENV_ROOT/bin:/root/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

ENV CONFIG_FILE=./conf.py

ENV PYTHON_VERSION 3.9.17

ENV POETRY_VERSION 1.5.1

WORKDIR /app

RUN apt update && DEBIAN_FRONTEND=noninteractive apt install -y \
    make build-essential libssl-dev zlib1g-dev libbz2-dev \
    libreadline-dev libsqlite3-dev wget curl llvm \
    libncurses5-dev libncursesw5-dev xz-utils tk-dev \
    libffi-dev liblzma-dev gcc libxml2-dev libxslt1-dev locales \
    git

RUN set -ex \
    && git clone --depth=1 https://github.com/pyenv/pyenv.git /root/.pyenv \
    && pyenv install $PYTHON_VERSION \
    && pyenv global $PYTHON_VERSION \
    && pyenv rehash \
    && pip install pipx \
    && pipx install poetry==$POETRY_VERSION

RUN echo 'en_AU.UTF-8 UTF-8' > /etc/locale.gen && /usr/sbin/locale-gen

RUN mkdir -p /app/files/transactions /app/files/uploads

COPY . .

RUN poetry install --no-dev

COPY migrations ./migrations

COPY Procfile boot.sh ./

RUN chmod a+x boot.sh

ENTRYPOINT ["./boot.sh"]

EXPOSE 5000
