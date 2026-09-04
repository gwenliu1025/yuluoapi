package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestBillingServiceConvertsOfficialUSDCardToCNY(t *testing.T) {
	pricingService := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"fixture-model": {
			InputCostPerToken:           2e-6,
			OutputCostPerToken:          8e-6,
			CacheReadInputTokenCost:     0.2e-6,
			OutputCostPerImage:          0.1,
			OutputCostPerImageToken:     3e-6,
			InputCostPerImageToken:      1e-6,
			CacheCreationInputTokenCost: 2.5e-6,
		},
	}}
	svc := NewBillingService(&config.Config{}, pricingService)

	pricing, err := svc.GetModelPricing("fixture-model")
	require.NoError(t, err)
	require.InDelta(t, usdToCNY(2e-6), pricing.InputPricePerToken, 1e-12)
	require.InDelta(t, usdToCNY(8e-6), pricing.OutputPricePerToken, 1e-12)
	require.InDelta(t, usdToCNY(0.2e-6), pricing.CacheReadPricePerToken, 1e-12)
	require.InDelta(t, usdToCNY(2.5e-6), pricing.CacheCreationPricePerToken, 1e-12)
	require.InDelta(t, usdToCNY(1e-6), pricing.ImageInputPricePerToken, 1e-12)
	require.InDelta(t, usdToCNY(3e-6), pricing.ImageOutputPricePerToken, 1e-12)
}

func TestConfiguredChannelPriceIsAlreadyCNY(t *testing.T) {
	svc := NewBillingService(&config.Config{}, &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"fixture-model": {InputCostPerToken: 2e-6, OutputCostPerToken: 8e-6},
	}})
	input := 1.95e-6
	output := 5.85e-6

	pricing, err := svc.GetModelPricingWithChannel("fixture-model", &ChannelModelPricing{
		InputPrice:  &input,
		OutputPrice: &output,
	})
	require.NoError(t, err)
	require.Equal(t, input, pricing.InputPricePerToken)
	require.Equal(t, output, pricing.OutputPricePerToken)
}
