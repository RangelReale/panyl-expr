package panylexpr

import (
	"context"
	"testing"
	"time"

	"github.com/RangelReale/panyl/v2"
	"github.com/stretchr/testify/assert"
)

func TestCondition(t *testing.T) {
	ctx := context.Background()
	e, err := NewCondition(&Plugin{}, `metadata.message == "incoming request" && int(data["http-status"]) >= 300 && int(data["http-status"]) <= 399`,
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

	err = e.Process(ctx, pp)
	assert.NoError(t, err)
	assert.Equal(t, panyl.MetadataLevelWARNING, pp.Metadata[panyl.MetadataLevel])
	assert.Equal(t, "1", pp.Data["a"])
}

func TestCondition2(t *testing.T) {
	ctx := context.Background()
	e, err := NewCondition(&Plugin{}, `"http-status" in data && int(data["http-status"]) >= 300 && int(data["http-status"]) <= 399`,
		`set_metadata(MetadataLevel, MetadataLevelWARNING)`)
	assert.NoError(t, err)
	pp := &panyl.Item{
		Metadata: map[string]any{
			panyl.MetadataTimestamp: time.Now(),
			panyl.MetadataMessage:   "incoming request",
			panyl.MetadataLevel:     panyl.MetadataLevelINFO,
		},
		Data: map[string]any{
			// "http-status": "302",
			"http-path": "/healthz",
		},
	}

	err = e.Process(ctx, pp)
	assert.NoError(t, err)
	assert.Equal(t, panyl.MetadataLevelINFO, pp.Metadata[panyl.MetadataLevel])
}

func TestCondition3(t *testing.T) {
	ctx := context.Background()
	e, err := NewCondition(&Plugin{}, `metadata.message == "incoming request" && int(data["http-status"]) >= 300 && int(data["http-status"]) <= 399`,
		`set_metadata(MetadataMessage, sprintf("%s [%s]", metadata.message, data["http-status"]))`)
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

	err = e.Process(ctx, pp)
	assert.NoError(t, err)
	assert.Equal(t, "incoming request [302]", pp.Metadata[panyl.MetadataMessage])
}
