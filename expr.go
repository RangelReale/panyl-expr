package panylexpr

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"

	"github.com/RangelReale/panyl/v2"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger     *slog.Logger
	Conditions []Condition
	Constants  map[string]any
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

func (e *Config) AddConstants(c map[string]any) {
	maps.Copy(e.Constants, c)
}

func (e *Config) Process(ctx context.Context, item *panyl.Item) error {
	for _, condition := range e.Conditions {
		err := condition.Process(ctx, e, item)
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
		cc, err := loadConditionConfig(e, r)
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
		cc, err := loadConditionConfig(e, f)
		if err != nil {
			return fmt.Errorf("error decoding config: %v", err)
		}
		e.Conditions = append(e.Conditions, cc...)
		return nil
	}
}

func WithConfigConstants(constants map[string]any) ConfigOption {
	return func(e *Config) error {
		if e.Constants == nil {
			e.Constants = map[string]any{}
		}
		maps.Copy(e.Constants, constants)
		return nil
	}
}

func loadConditionConfig(cfg *Config, r io.Reader) ([]Condition, error) {
	var cc ConditionConfig

	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	err := dec.Decode(&cc)
	if err != nil {
		return nil, fmt.Errorf("error decoding config: %v", err)
	}

	var ret []Condition
	for _, c := range cc.Conditions {
		cond, err := NewCondition(cfg, c.When, c.Do)
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
