package panylexpr

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/RangelReale/panyl/v2"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type Condition struct {
	When *vm.Program
	Do   *vm.Program
}

func NewCondition(config *Config, when, do string) (Condition, error) {
	whenEnv := maps.Clone(defaultWhenEnv)
	maps.Copy(whenEnv, config.Constants)
	doEnv := maps.Clone(defaultDoEnv)
	maps.Copy(doEnv, config.Constants)

	programWhen, err := expr.Compile(when,
		expr.AsBool(),
		expr.Env(whenEnv),
	)
	if err != nil {
		return Condition{}, fmt.Errorf("error parsing '%s': %w", when, err)
	}
	programDo, err := expr.Compile(do,
		expr.Env(doEnv),
		expr.AsBool())
	if err != nil {
		return Condition{}, fmt.Errorf("error parsing '%s': %w", do, err)
	}
	return Condition{
		When: programWhen,
		Do:   programDo,
	}, nil
}

func (e Condition) Process(config *Config, item *panyl.Item) error {
	condEnv := map[string]any{
		"metadata": item.Metadata,
		"data":     item.Data,
		"line":     item.Line,
		"source":   item.Source,
		"source_json": func(name string) (any, error) {
			return getJSONField(item.Source, name)
		},
		"log": func(level string, message string) (bool, error) {
			return log(config.Logger, level, message)
		},
	}
	maps.Copy(condEnv, defaultEnv)
	maps.Copy(condEnv, config.Constants)

	output, err := expr.Run(e.When, condEnv)
	if err != nil {
		return err
	}
	if !output.(bool) {
		return nil
	}

	resultEnv := map[string]any{
		"set_data": func(name, value string) bool {
			item.Data[name] = value
			return true
		},
		"add_data_list": func(name, value string) bool {
			item.Data.ListValueAdd(name, value)
			return true
		},
		"set_metadata": func(name, value string) bool {
			item.Metadata[name] = value
			return true
		},
		"add_metadata_list": func(name, value string) bool {
			item.Metadata.ListValueAdd(name, value)
			return true
		},
		"set_source": func(source string) bool {
			item.Source = source
			return true
		},
		"set_source_json": func(name string, value any) (bool, error) {
			src, err := setJSONField(item.Source, name, value)
			if err != nil {
				return false, err
			}
			item.Source = src
			return true, nil
		},
	}
	maps.Copy(resultEnv, condEnv)
	result, err := expr.Run(e.Do, resultEnv)
	if err != nil {
		return err
	}
	if !result.(bool) {
		return errors.New("do returned false")
	}
	return nil
}

func getJSONField(source string, name string) (any, error) {
	src := map[string]any{}
	err := json.Unmarshal([]byte(source), &src)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling source as JSON: %w", err)
	}
	if v, ok := src[name]; ok {
		return v, nil
	}
	return nil, nil
}

func setJSONField(source string, name string, value any) (string, error) {
	src := map[string]any{}
	err := json.Unmarshal([]byte(source), &src)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling source as JSON: %w", err)
	}
	src[name] = value
	enc, err := json.Marshal(src)
	if err != nil {
		return "", fmt.Errorf("error marshalling source back to JSON: %w", err)
	}
	return string(enc), nil
}

var defaultEnv = map[string]any{
	"MetadataTimestamp":        panyl.MetadataTimestamp,
	"MetadataMessage":          panyl.MetadataMessage,
	"MetadataLevel":            panyl.MetadataLevel,
	"MetadataApplication":      panyl.MetadataApplication,
	"MetadataCategory":         panyl.MetadataCategory,
	"MetadataOriginalCategory": panyl.MetadataOriginalCategory,
	"MetadataSkip":             panyl.MetadataSkip,

	"MetadataLevelTRACE":   panyl.MetadataLevelTRACE,
	"MetadataLevelDEBUG":   panyl.MetadataLevelDEBUG,
	"MetadataLevelINFO":    panyl.MetadataLevelINFO,
	"MetadataLevelWARNING": panyl.MetadataLevelWARNING,
	"MetadataLevelERROR":   panyl.MetadataLevelERROR,
}

var defaultWhenEnv = map[string]any{
	"metadata":    map[string]any{},
	"data":        map[string]any{},
	"line":        "",
	"source":      "",
	"source_json": func(name string) (any, error) { return nil, nil },
	"log":         func(level string, message string) (bool, error) { return true, nil },
}

var defaultDoEnv = map[string]any{
	"set_data":          func(name, value string) bool { return true },
	"add_data_list":     func(name, value string) bool { return true },
	"set_metadata":      func(name, value string) bool { return true },
	"add_metadata_list": func(name, value string) bool { return true },
	"set_source":        func(name, value string) bool { return true },
	"set_source_json":   func(name, value string) (bool, error) { return true, nil },
}

func init() {
	maps.Copy(defaultDoEnv, defaultWhenEnv)
	maps.Copy(defaultDoEnv, defaultEnv)
	maps.Copy(defaultWhenEnv, defaultEnv)
}
