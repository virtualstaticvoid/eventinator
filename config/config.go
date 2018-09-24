package config

import (
	"flag"
	"fmt"

	"github.com/google/uuid"
)

const (
	defaultNATSServer         = "nats://localhost:4222"
	defaultClusterId          = "test-cluster"
	defaultCertificateFile    = "server.pem"
	defaultCertificateKeyFile = "server.key"
)

type Configuration struct {
	ServerURL          string
	ClusterId          string
	ClientId           string
	VerboseLogging     bool
	Port               int
	MetricsEnabled     bool
	MetricsPort        int
	Secure             bool
	CertificateFile    string
	CertificateKeyFile string
}

func randomClientId() string {
	u, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	clientID := fmt.Sprintf("e8r-%s", u.String())
	return clientID
}

func ParseCommandLine() *Configuration {
	cfg := Configuration{}
	flag.StringVar(&cfg.ServerURL, "server", defaultNATSServer, "Comma separated list of NATS URIs to connect to")
	flag.StringVar(&cfg.ClusterId, "clusterID", defaultClusterId, "NATS Cluser ID to connect to")
	flag.StringVar(&cfg.ClientId, "clientID", randomClientId(), "Client ID used to identify the instance")
	flag.BoolVar(&cfg.VerboseLogging, "verbose", false, "Whether or not to enable verbose logging")
	flag.IntVar(&cfg.Port, "port", 5300, "Port to bind to")
	flag.BoolVar(&cfg.MetricsEnabled, "metricsEnabled", true, "Whether or not the metrics scrape endpoint is enabled")
	flag.IntVar(&cfg.MetricsPort, "metricsPort", 9000, "Port to bind to for metrics scrapers")
	flag.BoolVar(&cfg.Secure, "secure", false, "Whether or not to use TLS")
	flag.StringVar(&cfg.CertificateFile, "cert", defaultCertificateFile, "TLS certificate file")
	flag.StringVar(&cfg.CertificateKeyFile, "key", defaultCertificateKeyFile, "TLS private key file")
	flag.Parse()
	return &cfg
}
