package panylexpr

import (
	"testing"
	"time"

	"github.com/RangelReale/panyl/v2"
	"github.com/stretchr/testify/assert"
)

func TestCondition(t *testing.T) {
	e, err := NewCondition(`metadata.message == "incoming request" && int(data["http-status"]) >= 300 && int(data["http-status"]) <= 399`,
		`set_metadata(MetadataLevel, MetadataLevelWARNING) && set_data("a", "1")`)
	assert.NoError(t, err)
	pp := &panyl.Item{
		Metadata: map[string]any{
			panyl.MetadataTimestamp: time.Now(),
			panyl.MetadataMessage:   "incoming request",
			panyl.MetadataLevel:     panyl.MetadataLevelINFO,
		},
		Data: map[string]any{
			"http-status": "302",
			"http-path":   "/healthz",
		},
	}

	err = e.Process(&Config{}, pp)
	assert.NoError(t, err)
	assert.Equal(t, panyl.MetadataLevelWARNING, pp.Metadata[panyl.MetadataLevel])
	assert.Equal(t, "1", pp.Data["a"])
}
