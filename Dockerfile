FROM alpine:3.7

ENV FLASK_APP=penny.py

ENV CONFIG_FILE=conf.py

WORKDIR /app

RUN mkdir -p /app/files/transactions /app/files/uploads

RUN apk upgrade --update-cache --available

RUN apk add python3 python3-dev musl-dev py3-pip gcc libffi-dev openssl-dev libxml2-dev libxslt-dev

COPY . .

RUN pip3 install -r /app/requirements.txt

CMD flask run --host=0.0.0.0 --port=5000

EXPOSE 5000
