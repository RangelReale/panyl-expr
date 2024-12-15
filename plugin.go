package panylexpr

import (
	"context"

	"github.com/RangelReale/panyl/v2"
)

type Expr struct {
	Config *Config
}

var _ panyl.PluginPostProcess = (*Expr)(nil)

func New(options ...ConfigOption) (*Expr, error) {
	cfg, err := NewConfig(options...)
	if err != nil {
		return nil, err
	}
	return &Expr{Config: cfg}, nil
}

func (e Expr) PostProcessOrder() int {
	return 10
}

func (e Expr) PostProcess(ctx context.Context, result *panyl.Item) (bool, error) {
	if e.Config == nil {
		return false, nil
	}
	return false, e.Config.Process(result)
}

func (e Expr) IsPanylPlugin() {}
