list:
	@echo ""
	@echo "Useful targets:"
	@echo ""
	@echo "  - 'make init' > run initialization process to set envs"
	@echo "  - 'make install' > run installation of the daemon"
	@echo "  - 'image.build' > docker build app"
	@echo "  - 'image.push' > docker push app"

init:
	export AMQP_CONNECTION=amqp://rabbit:Qwerty11@rabbitmq-dev1.gm:5672/
	export AMQP_EXCHANGE=fanout_services
	export AMQP_QUEUE=beta.upload.html_pages_beta
	export AMQP_EXCHANGE_TYPE=fanout
	export AMQP_HTML_PATH=/mnt/static/html/

install:
	go build -o html-render main.go

image.build:
	docker build -t docker-proxy.gismeteo.net/microservice/html-render/app .
image.push:
	docker push docker-proxy.gismeteo.net/microservice/html-render/app

