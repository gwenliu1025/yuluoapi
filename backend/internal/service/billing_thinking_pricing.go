package service

import "strings"

// requestModePricingForContext 按官方同档的思考/显式缓存价差沿用对应计费项折扣，
// 不从上游成本反推售价，输入与缓存创建价格保持配置值。
func (s *BillingService) requestModePricingForContext(model string, pricing *ModelPricing, contextTokens int, thinking, explicitCache bool) *ModelPricing {
	if pricing == nil {
		return nil
	}
	official, err := s.GetModelPricing(model)
	if err != nil || official == nil {
		return pricing
	}
	if interval := FindMatchingInterval(official.Intervals, contextTokens); interval != nil {
		official = intervalToModelPricing(interval, official, nil)
	}
	cloned := *pricing
	if thinking && official.ThinkingOutputPricePerToken != nil && official.OutputPricePerToken > 0 {
		cloned.OutputPricePerToken *= *official.ThinkingOutputPricePerToken / official.OutputPricePerToken
	}
	if explicitCache && official.ExplicitCacheReadPricePerToken != nil && official.CacheReadPricePerToken > 0 {
		cloned.CacheReadPricePerToken *= *official.ExplicitCacheReadPricePerToken / official.CacheReadPricePerToken
	}
	return &cloned
}

func reasoningEffortEnablesThinking(effort *string) bool {
	if effort == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(*effort)) {
	case "", "none", "disabled":
		return false
	default:
		return true
	}
}
