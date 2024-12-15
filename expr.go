package panylexpr

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/RangelReale/panyl/v2"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger     *slog.Logger
	Conditions []Condition
}

func NewConfig(options ...ConfigOption) (*Config, error) {
	ret := &Config{}
	for _, opt := range options {
		if err := opt(ret); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (e *Config) Process(item *panyl.Item) error {
	for _, condition := range e.Conditions {
		err := condition.Process(e, item)
		if err != nil {
			return err
		}
	}
	return nil
}

type ConfigOption func(*Config) error

func WithConfigLogger(logger *slog.Logger) ConfigOption {
	return func(e *Config) error {
		e.Logger = logger
		return nil
	}
}

func WithConfigReader(r io.Reader) ConfigOption {
	return func(e *Config) error {
		cc, err := loadConditionConfig(r)
		if err != nil {
			return fmt.Errorf("error decoding config: %v", err)
		}
		e.Conditions = append(e.Conditions, cc...)
		return nil
	}
}

func WithConfigFile(filename string) ConfigOption {
	return func(e *Config) error {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		cc, err := loadConditionConfig(f)
		if err != nil {
			return fmt.Errorf("error decoding config: %v", err)
		}
		e.Conditions = append(e.Conditions, cc...)
		return nil
	}
}

func loadConditionConfig(r io.Reader) ([]Condition, error) {
	var cc ConditionConfig

	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	err := dec.Decode(&cc)
	if err != nil {
		return nil, fmt.Errorf("error decoding config: %v", err)
	}

	var ret []Condition
	for _, c := range cc.Conditions {
		cond, err := NewCondition(c.When, c.Do)
		if err != nil {
			return nil, err
		}
		ret = append(ret, cond)
	}

	return ret, nil
}

type ConditionConfig struct {
	Conditions []ConditionItemConfig `yaml:"conditions"`
}

type ConditionItemConfig struct {
	When string `yaml:"when"`
	Do   string `yaml:"do"`
}
