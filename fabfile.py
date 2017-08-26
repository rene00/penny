from fabric.api import local, env
from fabric.decorators import task

env.flask_host = '127.0.0.1'
env.flask_port = 5000

@task
def run():
    build()
    local("mkdir -p files/transactions files/uploads")
    local('FLASK_APP=penny.py CONFIG_FILE=conf.py \
            flask run \
                --host={0.flask_host} \
                --port={0.flask_port}'.
            format(env))

@task
def run_queue():
    local('CONFIG_FILE=conf.py \
            rqworker \
            --url redis://localhost:6379/0 \
            --verbose \
            --path=.')

@task
def build():
    local('pip install -r requirements.txt')

