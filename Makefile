#!/usr/bin/make

DEFAULT_PYTHON_VERSION ?= 2.7
VENV = pyvenv-$(DEFAULT_PYTHON_VERSION)
FLASK_DEBUG ?= 0
FLASK_HOST ?= 127.0.0.1
FLASK_PORT ?= 5000

venv: requirements.txt
	rm -rf $@
	virtualenv $@ -p python$(DEFAULT_PYTHON_VERSION)
	$@/bin/pip install -r requirements.txt

update: 
	venv/bin/pip install -r requirements.txt

clean: clean_pyc
	rm -rf venv

clean_pyc:
	find . -name '*.pyc' -exec rm -f {} +
	find . -name '*.pyo' -exec rm -f {} +

db_migrate: venv
	CONFIG_FILE=conf.py \
	venv/bin/python manage.py db upgrade

run_www: venv
	mkdir -p files/transactions files/uploads
	FLASK_APP=penny.py \
	CONFIG_FILE=conf.py \
	FLASK_DEBUG=$(FLASK_DEBUG) \
	venv/bin/flask run --host=$(FLASK_HOST) --port=$(FLASK_PORT)

run_queue: venv
	CONFIG_FILE=conf.py \
	    venv/bin/rqworker \
		--url redis://localhost:6379/0 \
		--verbose \
		--path=.
