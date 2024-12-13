package panylexpr

import (
	"errors"
	"maps"

	"github.com/RangelReale/panyl"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type Expression struct {
	conditions []Condition
}

type Condition struct {
	When *vm.Program
	Do   *vm.Program
}

func NewExpression() (*Expression, error) {
	whenEnv := maps.Clone(defaultWhenEnv)
	maps.Copy(whenEnv, defaultEnv)
	doEnv := maps.Clone(defaultDoEnv)
	maps.Copy(doEnv, defaultEnv)

	programWhen, err := expr.Compile(`metadata.message == "incoming request" && int(data["http-status"]) >= 300 && int(data["http-status"]) <= 399`,
		expr.AsBool(),
		expr.Env(whenEnv),
	)
	if err != nil {
		return nil, err
	}
	programDo, err := expr.Compile(`set_metadata(Metadata_Level, MetadataLevel_WARNING) && set_data("a", "1")`,
		expr.Env(doEnv),
		expr.AsBool())
	if err != nil {
		return nil, err
	}
	return &Expression{
		conditions: []Condition{
			{When: programWhen, Do: programDo},
		},
	}, nil
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

func (e *Expression) Process(p *panyl.Process) error {
	for _, condition := range e.conditions {
		condEnv := map[string]any{
			"metadata": p.Metadata,
			"data":     p.Data,
		}
		maps.Copy(condEnv, defaultEnv)

		output, err := expr.Run(condition.When, condEnv)
		if err != nil {
			return err
		}
		if !output.(bool) {
			continue
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
		result, err := expr.Run(condition.Do, resultEnv)
		if err != nil {
			return err
		}
		if !result.(bool) {
			return errors.New("do returned false")
		}
	}
	return nil
}
