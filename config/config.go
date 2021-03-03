package config

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	defaultAddress        = ":8443"
	defaultMetricsAddress = ":8080"
)

type Config struct {
	Address        string
	MetricsAddress string
	CertFile       string
	Logger         micrologger.Logger
	KeyFile        string
}

func Parse() (Config, error) {
	var err error
	var config Config

	// Create a new logger that is used by all admitters.
	var newLogger micrologger.Logger
	{
		newLogger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
		config.Logger = newLogger
	}

	kingpin.Flag("address", "The address to listen on").Default(defaultAddress).StringVar(&config.Address)
	kingpin.Flag("metrics-address", "The metrics address for Prometheus").Default(defaultMetricsAddress).StringVar(&config.MetricsAddress)
	//kingpin.Flag("tls-cert-file", "File containing the certificate for HTTPS").Required().StringVar(&config.CertFile)
	//kingpin.Flag("tls-key-file", "File containing the private key for HTTPS").Required().StringVar(&config.KeyFile)

	kingpin.Parse()

	return config, nil
}
