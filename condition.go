package panylexpr

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/RangelReale/panyl"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type Condition struct {
	When *vm.Program
	Do   *vm.Program
}

func NewCondition(when, do string) (Condition, error) {
	programWhen, err := expr.Compile(when,
		expr.AsBool(),
		expr.Env(defaultWhenEnv),
	)
	if err != nil {
		return Condition{}, fmt.Errorf("error parsing '%s': %w", when, err)
	}
	programDo, err := expr.Compile(do,
		expr.Env(defaultDoEnv),
		expr.AsBool())
	if err != nil {
		return Condition{}, fmt.Errorf("error parsing '%s': %w", do, err)
	}
	return Condition{
		When: programWhen,
		Do:   programDo,
	}, nil
}

func (e Condition) Process(config *Config, p *panyl.Process) error {
	condEnv := map[string]any{
		"metadata": p.Metadata,
		"data":     p.Data,
		"line":     p.Line,
		"source":   p.Source,
		"source_json": func(name string) (any, error) {
			return getJSONField(p.Source, name)
		},
		"log": func(level string, message string) (bool, error) {
			return log(config.Logger, level, message)
		},
	}
	maps.Copy(condEnv, defaultEnv)

	output, err := expr.Run(e.When, condEnv)
	if err != nil {
		return err
	}
	if !output.(bool) {
		return nil
	}

	resultEnv := map[string]any{
		"set_metadata": func(name, value string) bool {
			p.Metadata[name] = value
			return true
		},
		"set_data": func(name, value string) bool {
			p.Data[name] = value
			return true
		},
		"set_source": func(source string) bool {
			p.Source = source
			return true
		},
		"set_source_json": func(name string, value any) (bool, error) {
			src, err := setJSONField(p.Source, name, value)
			if err != nil {
				return false, err
			}
			p.Source = src
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
	"Metadata_Timestamp":        panyl.Metadata_Timestamp,
	"Metadata_Message":          panyl.Metadata_Message,
	"Metadata_Level":            panyl.Metadata_Level,
	"Metadata_Application":      panyl.Metadata_Application,
	"Metadata_Category":         panyl.Metadata_Category,
	"Metadata_OriginalCategory": panyl.Metadata_OriginalCategory,
	"Metadata_Skip":             panyl.Metadata_Skip,

	"MetadataLevel_TRACE":    panyl.MetadataLevel_TRACE,
	"MetadataLevel_DEBUG":    panyl.MetadataLevel_DEBUG,
	"MetadataLevel_INFO":     panyl.MetadataLevel_INFO,
	"MetadataLevel_WARNING":  panyl.MetadataLevel_WARNING,
	"MetadataLevel_ERROR":    panyl.MetadataLevel_ERROR,
	"MetadataLevel_CRITICAL": panyl.MetadataLevel_CRITICAL,
	"MetadataLevel_FATAL":    panyl.MetadataLevel_FATAL,
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
	"set_data":        func(name, value string) bool { return true },
	"set_metadata":    func(name, value string) bool { return true },
	"set_source":      func(name, value string) bool { return true },
	"set_source_json": func(name, value string) (bool, error) { return true, nil },
}

func init() {
	maps.Copy(defaultDoEnv, defaultWhenEnv)
	maps.Copy(defaultDoEnv, defaultEnv)
	maps.Copy(defaultWhenEnv, defaultEnv)
}
