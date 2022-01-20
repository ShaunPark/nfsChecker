package rabbitmq

import (
	"fmt"
	"strings"

	"github.com/ShaunPark/nfsMonitor/types"

	"github.com/streadway/amqp"
	"k8s.io/klog/v2"
)

type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	// sChannel   *amqp.Channel
	queueName  string
	msgChannel <-chan amqp.Delivery
	sendFunc   func(<-chan amqp.Delivery)
}

func (r RabbitMQClient) GetConn() *amqp.Connection {
	return r.conn
}

func NewRabbitMQClient(config types.RabbitMQ, sendFunc func(<-chan amqp.Delivery)) (*RabbitMQClient, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Server.Id, config.Server.Password, config.Server.Host, config.Server.Port)
	conn, err := amqp.DialConfig(url, amqp.Config{Vhost: config.VHost})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &RabbitMQClient{conn: conn, queueName: config.Queue, sendFunc: sendFunc}, nil
}

func failOnError(err error, msg string) {
	if err != nil {
		klog.Errorf("%s: %s", msg, err)
	}
}

func queueDeclare(queueName string, ch *amqp.Channel) amqp.Queue {
	arg := amqp.Table{}
	arg["x-message-ttl"] = 20000

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		arg,       // arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q
}

func (r *RabbitMQClient) SendResponse(queueName *string, coId *string, msg []byte) {
	klog.Infof("SendResponse %s", *queueName)
	ch, err := r.conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	if !strings.HasPrefix(*queueName, "amq.gen") {
		queueDeclare(*queueName, ch)
	}

	err = ch.Publish(
		"",         // exchange
		*queueName, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: *coId,
			Body:          msg,
		})
	failOnError(err, "Failed to publish a message")
}

func (r *RabbitMQClient) Close() {
	klog.Info("Close rabbitMQ channel")
	r.channel.Close()
	r.conn.Close()
}

func (r *RabbitMQClient) Run(stopCh <-chan interface{}) {
	ch, err := r.conn.Channel()
	r.channel = ch
	failOnError(err, "Failed to open a channel")

	q := queueDeclare(r.queueName, ch)

	r.msgChannel, err = r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	r.sendFunc(r.msgChannel)
}
