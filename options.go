package panylexpr

import (
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"

	"gopkg.in/yaml.v3"
)

type Option func(*Plugin) error

// WithLogger sets a logger for debugging purposes.
func WithLogger(logger *slog.Logger) Option {
	return func(e *Plugin) error {
		e.Logger = logger
		return nil
	}
}

// WithConfigReader sets an io.Reader to read the configuration file.
func WithConfigReader(r io.Reader) Option {
	return func(e *Plugin) error {
		cc, err := loadConditionConfig(e, r)
		if err != nil {
			return fmt.Errorf("error decoding config: %v", err)
		}
		e.Conditions = append(e.Conditions, cc...)
		return nil
	}
}

// WithConfigFile sets a filename to read the configuration file.
func WithConfigFile(filename string) Option {
	return func(e *Plugin) error {
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

// WithConstants adds extra constants to the scripts.
func WithConstants(constants map[string]any) Option {
	return func(e *Plugin) error {
		if e.Constants == nil {
			e.Constants = map[string]any{}
		}
		maps.Copy(e.Constants, constants)
		return nil
	}
}

// loadConditionConfig loads conditions from the configuration file.
func loadConditionConfig(cfg *Plugin, r io.Reader) ([]Condition, error) {
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
