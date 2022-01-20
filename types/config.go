package types

type Config struct {
	TestMode        bool     `yaml:"testMode"`
	ProcessInterval int      `yaml:"processInterval"`
	RabbitMQ        RabbitMQ `yaml:"rabbitmq"`
}

type RabbitMQ struct {
	Server                   Host   `yaml:"server"`
	VHost                    string `yaml:"vHost"`
	Queue                    string `yaml:"queueName"`
	UseReplyQueueFromMessage bool   `yaml:"useReplyQueueFromMessage"`
	ReplyQueue               string `yaml:"replyQueueName"`
}

type Host struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Id       string `yaml:"id"`
	Password string
}
