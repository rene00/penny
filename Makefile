#!/usr/bin/make

CONFIG_FILE ?= "./conf.py"
FLASK_DEBUG ?= 1
FLASK_HOST ?= 127.0.0.1
FLASK_PORT ?= 5000
FLASK_APP = penny

clean: 
	find . -name '*.pyc' -exec rm -f {} +
	find . -name '*.pyo' -exec rm -f {} +

run: run_www

install-deps:
	./scripts/install-deps.sh

run_www: 
	mkdir -p files/transactions files/uploads
	FLASK_DEBUG=$(FLASK_DEBUG) \
	CONFIG_FILE=$(CONFIG_FILE) \
	FLASK_APP=$(FLASK_APP) \
	poetry run flask run --host=$(FLASK_HOST) --port=$(FLASK_PORT)

run_queue: 
	CONFIG_FILE=$(CONFIG_FILE) \
	    poetry run rqworker \
		--url redis://localhost:6379/0 \
		--verbose \
		--path=.
