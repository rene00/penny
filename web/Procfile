redis: /usr/bin/redis-server
penny_www: poetry run flask run --host=0.0.0.0 --port=5000
penny_queue: poetry run rqworker --url redis://localhost:6379/0 --verbose --path=/app
