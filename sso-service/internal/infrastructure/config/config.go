package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

type app struct {
	Env     string `yaml:"env"`
	Version string `yaml:"version"`
	Name    string `yaml:"name"`
}

func (a *app) validate() error {
	if a.Env == "" {
		a.Env = "dev"
	}

	return nil
}

type server struct {
	GRPC struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Protocol string `yaml:"protocol"`
		TLS      struct {
			Enable bool `yaml:"enable"`
		} `yaml:"tls"`
	} `yaml:"grpc"`
	Timeout time.Duration `yaml:"timeout"`
}

func (s *server) validate() error {
	if s.GRPC.Port <= 0 {
		return errors.New("invalid port")
	}

	// The validity of protocol will be checked when the listener is initialized
	if s.GRPC.Protocol == "" {
		s.GRPC.Protocol = "tcp"
	}

	if s.Timeout < time.Millisecond {
		return errors.New("invalid timeout")
	}

	return nil
}

type jwt struct {
	TTL         time.Duration `yaml:"ttl"`
	PrivatePath string        `yaml:"private-key-path"`
	PublicPath  string        `yaml:"public-key-path"`
}

func (j *jwt) validate() error {
	if j.TTL < time.Millisecond {
		return errors.New("too little ttl")
	}
	if j.PrivatePath == "" {
		return errors.New("invalid path to private key")
	}
	if j.PublicPath == "" {
		return errors.New("invalid path to public key")
	}

	return nil
}

type mongo struct {
	Host string `yaml:"host"`
	Port string `yaml:"Port"`
}

func (m *mongo) validate() error {
	if m.Host == "" {
		m.Host = "localhost"
	}
	if m.Port == "" {
		m.Port = "27017"
	}

	return nil
}

type redis struct {
	Host       string        `yaml:"host"`
	Port       string        `yaml:"port"`
	Password   string        `yaml:"password"`
	RefreshTTL time.Duration `yaml:"refresh-ttl"`
}

func (r *redis) validate() error {
	if r.Host == "" {
		r.Host = "localhost"
	}
	if r.Port == "" {
		r.Port = "6379"
	}

	return nil
}

type Config struct {
	App     app    `yaml:"app"`
	Server  server `yaml:"server"`
	Secrets struct {
		JWT   jwt   `yaml:"jwt"`
		Mongo mongo `yaml:"mongo"`
		Redis redis `yaml:"redis"`
	} `yaml:"secrets"`
}

// Path to config file (may be get from .env)
func New(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	bytes = []byte(os.ExpandEnv(string(bytes)))

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustNew(path string) *Config {
	cfg, err := New(path)
	if err != nil {
		slog.Error("failed to parse config", "error", err.Error())
		os.Exit(1)
	}

	return cfg
}

func (c *Config) Validate() error {
	if err := c.App.validate(); err != nil {
		return fmt.Errorf("invalid app: %w", err)
	}
	if err := c.Server.validate(); err != nil {
		return fmt.Errorf("invalid server: %w", err)
	}
	if err := c.Secrets.JWT.validate(); err != nil {
		return fmt.Errorf("invalid jwt: %w", err)
	}
	if err := c.Secrets.Mongo.validate(); err != nil {
		return fmt.Errorf("invalid mongo: %w", err)
	}
	if err := c.Secrets.Redis.validate(); err != nil {
		return fmt.Errorf("invalid redis: %w", err)
	}

	return nil
}
