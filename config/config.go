package config

// Config for Application.
type Config struct {
	Service struct {
		Name    string `default:"service name"`
		Version string `default:"1.0.0" env:"KC_VERSION"`
		Address string `default:":3000"`
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
		Brokers  string `required:"true" default:"127.0.0.1:9092" env:"KAFKA_BROKERS"`
		Group    string `default:"km-default-consumer" env:"KAFKA_GROUP"`
		Version  string `default:"0.10.2" env:"KAFKA_VERSION"`
		Producer bool   `default:"false" env:"PRODUCER"`
		SSL      SSLConfig
	}
}

// SSLConfig for Kafka.
type SSLConfig struct {
	Enable                                    bool `default:"false"`
	ClientCertFile, ClientKeyFile, CACertFile string
}
