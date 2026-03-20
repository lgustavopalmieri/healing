package kafka

import (
	"crypto/tls"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
)

type AuthConfig struct {
	SASLMechanism string
	SASLUsername  string
	SASLPassword  string
	UseTLS        bool
}

func AuthOpts(cfg AuthConfig) ([]kgo.Opt, error) {
	var opts []kgo.Opt

	if cfg.UseTLS {
		opts = append(opts, kgo.DialTLSConfig(&tls.Config{}))
	}

	switch cfg.SASLMechanism {
	case "PLAIN":
		mechanism := plain.Auth{
			User: cfg.SASLUsername,
			Pass: cfg.SASLPassword,
		}.AsMechanism()
		opts = append(opts, kgo.SASL(mechanism))
	case "":
	default:
		return nil, fmt.Errorf("unsupported SASL mechanism: %s", cfg.SASLMechanism)
	}

	return opts, nil
}
