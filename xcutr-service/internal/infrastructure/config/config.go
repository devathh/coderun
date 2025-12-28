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

type features struct {
	ClickhouseEnable bool `yaml:"clickhouse-enable"`
}

type server struct {
	GRPC struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Protocol string `yaml:"protocol"`
		TLS      struct {
			Enable bool `yaml:"enable"`
		} `yaml:"tls"`
	} `yaml:"grpc"`
}

func (s *server) validate() error {
	if s.GRPC.Host == "" {
		s.GRPC.Host = "localhost"
	}
	if s.GRPC.Port == "" {
		s.GRPC.Port = "50052"
	}
	if s.GRPC.Protocol == "" {
		s.GRPC.Protocol = "tcp"
	}

	return nil
}

type docker struct {
	ImageGo     string `yaml:"image-go"`
	ImagePython string `yaml:"image-python"`
}

func (d *docker) validate() error {
	if d.ImageGo == "" {
		return errors.New("invalid go image")
	}
	if d.ImagePython == "" {
		return errors.New("invalid python image")
	}

	return nil
}

type log struct {
	BufSize int `yaml:"buf-size"`
}

func (l *log) validate() error {
	if l.BufSize < 1 {
		return errors.New("too little buf size")
	}

	return nil
}

type jwt struct {
	PublicKeyPath string `yaml:"public-key-path"`
}

func (j *jwt) validate() error {
	if j.PublicKeyPath == "" {
		return errors.New("invalid path to public key")
	}

	return nil
}

type clickhouse struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	Username string `yaml:"username"`
	Database string `yaml:"database"`
}

func (ch *clickhouse) validate() error {
	if ch.Host == "" {
		ch.Host = "localhost"
	}
	if ch.Port == "" {
		ch.Port = "9000"
	}
	if ch.Password == "" {
		return errors.New("password cannot be empty")
	}
	if ch.Username == "" {
		return errors.New("username cannot be empty")
	}
	if ch.Database == "" {
		return errors.New("invalid database")
	}

	return nil
}

type Config struct {
	App      app      `yaml:"app"`
	Features features `yaml:"features"`
	Server   server   `yaml:"server"`
	Secrets  struct {
		Docker     docker     `yaml:"docker"`
		JWT        jwt        `yaml:"jwt"`
		Clickhouse clickhouse `yaml:"clickhouse"`
	} `yaml:"secrets"`
	Service struct {
		MaxTimeout time.Duration `yaml:"max-timeout"`
		Log        log           `yaml:"log"`
	} `yaml:"service"`
}

func New(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	bytes = []byte(os.ExpandEnv(string(bytes)))

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustNew(path string) *Config {
	cfg, err := New(path)
	if err != nil {
		slog.Error("failed to must create config", "error", err.Error())
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
	if err := c.Secrets.Docker.validate(); err != nil {
		return fmt.Errorf("invalid docker: %w", err)
	}
	if err := c.Service.Log.validate(); err != nil {
		return fmt.Errorf("invalid log: %w", err)
	}
	if err := c.Secrets.JWT.validate(); err != nil {
		return fmt.Errorf("invalid jwt: %w", err)
	}
	if err := c.Secrets.Clickhouse.validate(); err != nil {
		return fmt.Errorf("invalid clickhouse: %w", err)
	}

	return nil
}
