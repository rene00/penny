#!/usr/bin/make

FLASK_DEBUG ?= 0
FLASK_HOST ?= 127.0.0.1
FLASK_PORT ?= 5000

build: 
	pip install --no-cache-dir -r requirements.txt

docker_build:
	docker build .

docker_run:
	docker run -dit --restart always --publish 5000:5000 \
		-e CONFIG_FILE=/penny/penny.conf.py \
		--mount source=penny,target=/penny --name penny penny

clean: 
	find . -name '*.pyc' -exec rm -f {} +
	find . -name '*.pyo' -exec rm -f {} +

migrate: 
	CONFIG_FILE=conf.py python manage.py db upgrade

run: run_www

run_www: 
	mkdir -p files/transactions files/uploads
	FLASK_APP=penny.py \
	FLASK_DEBUG=$(FLASK_DEBUG) \
	flask run --host=$(FLASK_HOST) --port=$(FLASK_PORT)

run_queue: 
	CONFIG_FILE=conf.py \
	    rqworker \
		--url redis://localhost:6379/0 \
		--verbose \
		--path=.
