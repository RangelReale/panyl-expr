package panylexpr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/RangelReale/panyl/v2"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type Condition struct {
	Plugin *Plugin
	When   *vm.Program
	Do     *vm.Program
}

func NewCondition(plugin *Plugin, when, do string) (Condition, error) {
	whenEnv := maps.Clone(defaultWhenEnv)
	maps.Copy(whenEnv, plugin.Constants)
	doEnv := maps.Clone(defaultDoEnv)
	maps.Copy(doEnv, plugin.Constants)

	programWhen, err := expr.Compile(when,
		expr.AsBool(),
		expr.Env(whenEnv),
		expr.WithContext("ctx"),
	)
	if err != nil {
		return Condition{}, fmt.Errorf("error parsing '%s': %w", when, err)
	}
	programDo, err := expr.Compile(do,
		expr.Env(doEnv),
		expr.AsBool(),
		expr.WithContext("ctx"))
	if err != nil {
		return Condition{}, fmt.Errorf("error parsing '%s': %w", do, err)
	}
	return Condition{
		Plugin: plugin,
		When:   programWhen,
		Do:     programDo,
	}, nil
}

func (e Condition) Process(ctx context.Context, item *panyl.Item) error {
	output, err := expr.Run(e.When, e.getWhenEnv(ctx, item))
	if err != nil {
		return err
	}
	if !output.(bool) {
		return nil
	}

	result, err := expr.Run(e.Do, e.getDoEnv(ctx, item))
	if err != nil {
		return err
	}
	if !result.(bool) {
		return errors.New("do returned false")
	}
	return nil
}

func (e Condition) getWhenEnv(ctx context.Context, item *panyl.Item) map[string]any {
	condEnv := map[string]any{
		"ctx":      ctx,
		"metadata": item.Metadata,
		"data":     item.Data,
		"line":     item.Line,
		"source":   item.Source,
		"source_json": func(name string) (any, error) {
			return getJSONField(item.Source, name)
		},
		"log": func(level string, message string) (bool, error) {
			return log(e.Plugin.Logger, level, message)
		},
	}
	maps.Copy(condEnv, defaultEnv)
	maps.Copy(condEnv, e.Plugin.Constants)
	return condEnv
}

func (e Condition) getDoEnv(ctx context.Context, item *panyl.Item) map[string]any {
	resultEnv := e.getWhenEnv(ctx, item)
	maps.Copy(resultEnv, map[string]any{
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
		"sprintf": fmt.Sprintf,
	})
	return resultEnv
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
	"ctx":         context.Background(),
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
	"sprintf":           func(format string, a ...any) string { return "" },
}

func init() {
	maps.Copy(defaultWhenEnv, defaultEnv)
	maps.Copy(defaultDoEnv, defaultWhenEnv)
}
