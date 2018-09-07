package amqp

type ProducerConfig struct {
	Exchange     string `env:"amqp_exchange" required:"true"`
	ExchangeType string `env:"amqp_exchange_type" required:"true"`
	RoutingKey   string `env:"amqp_routing_key" required:"true"`
	URI          string `env:"amqp_uri" required:"true"`
}

type ConsumerConfig struct {
	Exchange     string `env:"amqp_exchange" required:"true"`
	ExchangeType string `env:"amqp_exchange_type" required:"true"`
	RoutingKey   string `env:"amqp_routing_key" required:"true"`
	URI          string `env:"amqp_uri" required:"true"`
	ConsumerTag string `env:"amqp_consumer_tag" required:"true"`
	QueueName   string `env:"amqp_queue_name" required:"true"`
}
