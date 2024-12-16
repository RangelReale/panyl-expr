package panylexpr

import (
	"context"
	"log/slog"
	"maps"

	"github.com/RangelReale/panyl/v2"
)

type Plugin struct {
	Logger     *slog.Logger
	Conditions []Condition
	Constants  map[string]any
}

var _ panyl.PluginPostProcess = (*Plugin)(nil)

func New(options ...ConfigOption) (*Plugin, error) {
	ret := &Plugin{}
	for _, opt := range options {
		if err := opt(ret); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (p *Plugin) PostProcessOrder() int {
	return 10
}

func (p *Plugin) PostProcess(ctx context.Context, item *panyl.Item) (bool, error) {
	for _, condition := range p.Conditions {
		err := condition.Process(ctx, p, item)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (p *Plugin) AddConstants(c map[string]any) {
	maps.Copy(p.Constants, c)
}

func (p *Plugin) IsPanylPlugin() {}
