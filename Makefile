#!/usr/bin/make

CONFIG_FILE ?= "./conf.py"
FLASK_DEBUG ?= 1
FLASK_HOST ?= 127.0.0.1
FLASK_PORT ?= 5000
FLASK_APP = penny

build: 
	pip3 install --no-cache-dir -r requirements.txt

docker_build:
	docker build . -t rene00/penny:latest

docker_run:
	docker run -dit --restart always --publish 5000:5000 \
		-e CONFIG_FILE=/app/penny/conf.py \
		--mount source=penny,target=/penny --name penny rene00/penny:latest

clean: 
	find . -name '*.pyc' -exec rm -f {} +
	find . -name '*.pyo' -exec rm -f {} +

migrate: 
	CONFIG_FILE=conf.py python manage.py db upgrade

run: run_www

install-deps:
	./scripts/install-deps.sh

run_www: 
	mkdir -p files/transactions files/uploads
	FLASK_DEBUG=$(FLASK_DEBUG) \
	CONFIG_FILE=$(CONFIG_FILE) \
	FLASK_APP=$(FLASK_APP) \
	flask run --host=$(FLASK_HOST) --port=$(FLASK_PORT)

run_queue: 
	CONFIG_FILE=$(CONFIG_FILE) \
	    rqworker \
		--url redis://localhost:6379/0 \
		--verbose \
		--path=.
