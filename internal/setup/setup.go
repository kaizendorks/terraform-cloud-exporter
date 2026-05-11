package setup

import (
	"bufio"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecthomas/kong"

	tfe "github.com/hashicorp/go-tfe"
)

type CLI struct {
	Organizations         []string `short:"o" env:"TF_ORGANIZATIONS" placeholder:"ORG1,ORG2" help:"List of the Organization names to scrape from (Ommit to scrape all)."`
	APIToken              string   `short:"t" env:"TF_API_TOKEN" help:"User token for autheticating with the API."`
	APITokenFile          *os.File `placeholder:"/path/to/file" help:"File containing user token for autheticating with the API."`
	APIAddress            string   `placeholder:"https://app.terraform.io/" env:"TF_API_ADDRESS" help:"Terraform API address to scrape metrics from."`
	APIInsecureSkipVerify bool     `help:"Accept any certificate presented by the API."`
	ListenAddress         string   `default:"0.0.0.0:9100" env:"TF_LISTEN_ADDRESS" help:"Address to listen on for web interface and telemetry."`
	LogLevel              string   `default:"info" enum:"debug,info,warn,error" help:"Only log messages with the given severity or above. One of: [${enum}]"`
	LogFormat             string   `default:"logfmt" enum:"logfmt,json" help:"Output format of log messages. One of: [${enum}]"`
}

type Config struct {
	CLI
	Client tfe.Client
	Logger *slog.Logger
}

// NewConfig returns a new Config object that was initialized according to the CLI params.
func NewConfig() Config {
	config := Config{}
	kong.Parse(&config.CLI)
	config.setupLogger()
	config.setupClient()
	return config
}

func (c *Config) setupLogger() {
	var logLevel slog.Level
	switch c.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: logLevel, AddSource: true}

	if c.LogFormat == "json" {
		c.Logger = slog.New(slog.NewJSONHandler(os.Stderr, opts))
	} else {
		c.Logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
	}
}

func (c *Config) setupClient() {
	config := &tfe.Config{}

	if c.APITokenFile != nil {
		defer c.APITokenFile.Close()
		scanner := bufio.NewScanner(c.APITokenFile)
		scanner.Scan()
		config.Token = scanner.Text()
	} else if c.APIToken != "" {
		config.Token = c.APIToken
	} else {
		c.Logger.Error("Error creating tfe client", "err", "Missing API Token.")
		os.Exit(1)
	}

	if c.APIAddress != "" {
		config.Address = c.APIAddress
		c.Logger.Info("Overwritten Terraform API address", "address", c.APIAddress)
	}

	if c.APIInsecureSkipVerify {
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					MinVersion:         tls.VersionTLS12,
				},
			},
		}
		c.Logger.Warn("HTTP InsecureSkipVerify is enabled.")
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		c.Logger.Error("Error creating tfe client", "err", err)
		os.Exit(1)
	}
	c.Client = *client
}
