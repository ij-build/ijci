package amqpclient

type AMQPConfig struct {
	Exchange   string `env:"amqp_exchange" required:"true"`
	RoutingKey string `env:"amqp_routing_key" required:"true"`
	URI        string `env:"amqp_uri" required:"true"`
}

type ProducerConfig struct {
	AMQPConfig
}

type ConsumerConfig struct {
	AMQPConfig
	ConsumerTag string `env:"amqp_consumer_tag" required:"true"`
	QueueName   string `env:"amqp_queue_name" required:"true"`
}
