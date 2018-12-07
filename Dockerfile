FROM golang:alpine3.8
ENV PROJECT_PATH=/go/src/app

WORKDIR ${PROJECT_PATH}
COPY . .

RUN apk add --no-cache git
RUN cd ${PROJECT_PATH} && go get github.com/streadway/amqp && go build -o app main.go && cp app /usr/local/bin/app && chmod 755 /usr/local/bin/app
CMD ["app"]
