package panylexpr

import "github.com/RangelReale/panyl"

type Expr struct {
}

var _ panyl.PluginPostProcess = (*Expr)(nil)

func (e Expr) PostProcessOrder() int {
	return 10
}

func (e Expr) PostProcess(result *panyl.Process) (bool, error) {
	return false, nil
}

func (e Expr) IsPanylPlugin() {}
