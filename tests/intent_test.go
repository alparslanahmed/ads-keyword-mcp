package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alparslanahmed/keyword-market-mcp/internal/intent"
)

func TestClassifyIntent(t *testing.T) {
	cases := []struct {
		keyword string
		want    string
	}{
		{"buy espresso maker", intent.Commercial},
		{"espresso maker fiyat", intent.Commercial},
		{"espresso maker satın al", intent.Commercial},
		{"best espresso maker", intent.Comparison},
		{"espresso maker review", intent.Comparison},
		{"espresso maker en iyi", intent.Comparison},
		{"how to make espresso", intent.Informational},
		{"espresso nedir", intent.Informational},
		{"espresso maker problem", intent.ProblemAware},
		{"baş ağrısı nasıl geçer", intent.ProblemAware},
		{"https://example.com/espresso", intent.Navigational},
		{"brandname.com", intent.Navigational},
		{"blue widget gadget", intent.Unknown},
	}
	for _, c := range cases {
		assert.Equalf(t, c.want, intent.ClassifyIntent(c.keyword, "English"), "keyword %q", c.keyword)
	}
}

func TestHasBuyingModifier(t *testing.T) {
	assert.True(t, intent.HasBuyingModifier("cheap espresso maker"))
	assert.True(t, intent.HasBuyingModifier("espresso maker fiyat"))
	assert.True(t, intent.HasBuyingModifier("espresso maker near me"))
	assert.False(t, intent.HasBuyingModifier("espresso maker"))
}
