package panylexpr

import (
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

func (e Condition) Process(p *panyl.Process) error {
	condEnv := map[string]any{
		"metadata": p.Metadata,
		"data":     p.Data,
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
	}
	maps.Copy(resultEnv, defaultEnv)
	result, err := expr.Run(e.Do, resultEnv)
	if err != nil {
		return err
	}
	if !result.(bool) {
		return errors.New("do returned false")
	}
	return nil
}

var defaultEnv = map[string]any{
	"Metadata_Timestamp": panyl.Metadata_Timestamp,
	"Metadata_Message":   panyl.Metadata_Message,
	"Metadata_Level":     panyl.Metadata_Level,

	"MetadataLevel_TRACE":    panyl.MetadataLevel_TRACE,
	"MetadataLevel_DEBUG":    panyl.MetadataLevel_DEBUG,
	"MetadataLevel_INFO":     panyl.MetadataLevel_INFO,
	"MetadataLevel_WARNING":  panyl.MetadataLevel_WARNING,
	"MetadataLevel_ERROR":    panyl.MetadataLevel_ERROR,
	"MetadataLevel_CRITICAL": panyl.MetadataLevel_CRITICAL,
	"MetadataLevel_FATAL":    panyl.MetadataLevel_FATAL,
}

var defaultWhenEnv = map[string]any{
	"metadata": map[string]any{},
	"data":     map[string]any{},
}

var defaultDoEnv = map[string]any{
	"set_data":     func(name, value string) bool { return true },
	"set_metadata": func(name, value string) bool { return true },
}

func init() {
	maps.Copy(defaultWhenEnv, defaultEnv)
	maps.Copy(defaultDoEnv, defaultEnv)
}
