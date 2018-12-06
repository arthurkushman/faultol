package faultal

import (
	"strconv"
	"time"
	"log"
	"os"
	"io/ioutil"
	"github.com/streadway/amqp"
	"fmt"
	"encoding/json"
	"strings"
)

const (
	UrlSlice = 2
	MainPage = "index"
)

var (
	uri          = os.Getenv("FAULTOL_AMQP_CONNECTION")
	exchange     = os.Getenv("FAULTOL_AMQP_EXCHANGE")
	exchangeType = os.Getenv("FAULTOL_AMQP_EXCHANGE_TYPE")
	queue        = os.Getenv("FAULTOL_AMQP_QUEUE")
	bindingKey   = os.Getenv("FAULTOL_AMQP_BINDING_KEY")
	consumerTag  = os.Getenv("FAULTOL_AMQP_CONSUMER_TAG")
	lifetime     = os.Getenv("FAULTOL_AMQP_LIFETIME")
	htmlPath     = os.Getenv("FAULTOL_HTML_PATH")
	maxThreads   = os.Getenv("FAULTOL_MAX_THREADS")
)

func run() {
	c, err := NewConsumer(uri, exchange, exchangeType, queue, bindingKey, consumerTag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	lt, err := strconv.Atoi(lifetime)

	if lt > 0 {
		log.Printf("running for %s", lifetime)
		time.Sleep(time.Duration(lt) * time.Second)
	} else {
		log.Printf("running forever")
		select {}
	}

	log.Printf("shutting down")

	if err := c.Shutdown(); err != nil {
		log.Fatalf("error during shutdown: %s", err)
	}
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func NewConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	var err error

	log.Printf("dialing %q", amqpURI)
	c.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck
		true,       // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go handle(deliveries, c.done)

	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

type Message struct {
	Uri  string `json:"url"`
	Data string `json:"data"`
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	var uriPath string
	var message Message

	var i = 0
	for d := range deliveries {
		uriPath = ""
		message = Message{}

		json.Unmarshal(d.Body, &message)
		if message.Uri == "/" {
			uriPath = "/" + MainPage // in case of a main page write index.html file
			recursiveMkDir(htmlPath)
		} else {
			urlSlice := strings.Split(message.Uri, "/")
			if len(urlSlice) <= UrlSlice {
				urlSlice = strings.Split(message.Uri, "_")
			}

			// todo: override on buffer str concat in future
			for _, uriPart := range urlSlice[:len(urlSlice)-UrlSlice] {
				uriPath += uriPart + "/"
			}

			recursiveMkDir(htmlPath + uriPath)

			for _, uriPart := range urlSlice[len(urlSlice)-UrlSlice:len(urlSlice)-1] {
				uriPath += uriPart
			}
		}

		maxT, err := strconv.Atoi(maxThreads)
		check(err)

		if i > maxT { // ждем отработки N кол-ва threads - разгружаем CPU la
			time.Sleep(300 * time.Millisecond)
		}

		// плодим threads (внутри handler thread) для многопоточной записи файлов на диск
		// синхронно создаем директорию и асинхронно пишем файлы
		go writeFile(uriPath, message.Data)

		// ack deliveries
		d.Ack(false)
		i++
	}

	log.Printf("handle: deliveries channel closed")
	done <- nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func recursiveMkDir(path string) {
	err := os.MkdirAll(path, 0755)
	check(err)
}

func writeFile(uriPath, d string) {
	data := []byte(d)
	err := ioutil.WriteFile(htmlPath+uriPath+".html", data, 0644)
	check(err)
}
