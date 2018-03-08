redis: /usr/bin/redis-server
penny_www: /usr/local/bin/flask run --host=0.0.0.0 --port=5000
penny_queue: /usr/local/bin/rqworker --url redis://localhost:6379/0 --verbose --path=/app
