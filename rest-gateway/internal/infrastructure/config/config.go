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
	HTTP struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"http"`
	ReadTimeout  time.Duration `yaml:"read-timeout"`
	WriteTimeout time.Duration `yaml:"write-timeout"`
	IdleTimeout  time.Duration `yaml:"idle-timeout"`
}

func (s *server) validate() error {
	if s.HTTP.Host == "" {
		s.HTTP.Host = "localhost"
	}
	if s.HTTP.Port == "" {
		s.HTTP.Port = "9090"
	}
	if s.ReadTimeout < time.Millisecond {
		return errors.New("too little read timeout")
	}
	if s.WriteTimeout < time.Millisecond {
		return errors.New("too little write timeout")
	}
	if s.IdleTimeout < time.Second {
		return errors.New("too little idle timeout")
	}

	return nil
}

type coderunService struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (cs *coderunService) validate() error {
	if cs.Host == "" {
		return errors.New("invalid host")
	}
	if cs.Port == "" {
		return errors.New("invalid port")
	}

	return nil
}

type Config struct {
	App      app    `yaml:"app"`
	Server   server `yaml:"server"`
	Services struct {
		CoderunSSO coderunService `yaml:"coderun-sso"`
	} `yaml:"services"`
}

// The path of config file can be used from .env
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

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustNew(path string) *Config {
	cfg, err := New(path)
	if err != nil {
		slog.Error("failed to load config", "error", err.Error())
		os.Exit(1)
	}

	return cfg
}

func (c *Config) validate() error {
	if err := c.App.validate(); err != nil {
		return fmt.Errorf("invalid app: %w", err)
	}
	if err := c.Server.validate(); err != nil {
		return fmt.Errorf("invalid server: %w", err)
	}
	if err := c.Services.CoderunSSO.validate(); err != nil {
		return fmt.Errorf("invalid coderun-sso: %w", err)
	}

	return nil
}
