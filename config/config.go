package config

// Config for Application.
type Config struct {
	Service struct {
		Name    string `default:"service name"`
		Version string `default:"1.0.0" env:"KC_VERSION"`
		Address string `default:":3000"`
	}

	Producer struct {
		Enable  bool   `default:"false" env:"PRODUCER_ENABLE"`
		Version string `default:"1.0.0" env:"PRODUCER_VERSION"`
	}

	Profiler struct {
		Enable    bool   `default:"false"`
		Listening string `default:"0.0.0.0:6060"`
	}

	Contacts []struct {
		Name  string
		Email string `required:"true"`
	}

	Kafka struct {
		Brokers string `default:"127.0.0.1:9092" env:"KAFKA_BROKERS"`
		Group   string `default:"km-default-consumer" env:"KAFKA_GROUP"`
		SSL     SSLConfig
	}
}

// SSLConfig for Kafka.
type SSLConfig struct {
	Enable                                    bool `default:"false"`
	ClientCertFile, ClientKeyFile, CACertFile string
}
