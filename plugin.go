package panylexpr

import "github.com/RangelReale/panyl"

type Expr struct {
	Config *Config
}

var _ panyl.PluginPostProcess = (*Expr)(nil)

func (e Expr) PostProcessOrder() int {
	return 10
}

func (e Expr) PostProcess(result *panyl.Process) (bool, error) {
	if e.Config == nil {
		return false, nil
	}
	return false, e.Config.Process(result)
}

func (e Expr) IsPanylPlugin() {}
