package engine

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var validate *validator.Validate

type Config struct {
	Name     string            `yaml:"name"`
	Handler  string            `yaml:"handler" validate:"required"`
	Env      map[string]string `yaml:"env"`
	Engine   string            `yaml:"engine" default:"executable"`
	Triggers []string          `yaml:"triggers" default"[\".\"]"`
	Channels []string          `yaml:"channels"`
	Threaded bool              `yaml:"threaded" default:"true"`
}

func ConfigInit() {
	validate = validator.New()
}

func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config

	if err = defaults.Set(&cfg); err != nil {
		return nil, err
	}

	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	// Validate config
	if err = validate.Struct(cfg); err != nil {
		return nil, err
	}

	if cfg.Name == "" {
		// Use config filename with extension stripped as name
		cfg.Name = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	}

	logrus.Debugf("Loaded config: %+v", cfg)
	return &cfg, nil
}
