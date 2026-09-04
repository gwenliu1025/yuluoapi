package service

// BillingCurrency 是雨落 API 余额、额度、售价、扣费和统计的唯一业务币种。
const BillingCurrency = "CNY"

// BillingUSDToCNYRate 仅用于把 LiteLLM 和内置官方美元价卡规范化为人民币。
// 运营侧显式配置的分组、渠道和套餐价格已经是人民币，不经过此换算。
const BillingUSDToCNYRate = 6.723579635

func usdToCNY(value float64) float64 {
	return value * BillingUSDToCNYRate
}

func pricingUSDToCNY(pricing *ModelPricing) *ModelPricing {
	if pricing == nil {
		return nil
	}
	converted := *pricing
	converted.InputPricePerToken = usdToCNY(converted.InputPricePerToken)
	converted.InputPricePerTokenPriority = usdToCNY(converted.InputPricePerTokenPriority)
	converted.ImageInputPricePerToken = usdToCNY(converted.ImageInputPricePerToken)
	converted.OutputPricePerToken = usdToCNY(converted.OutputPricePerToken)
	converted.OutputPricePerTokenPriority = usdToCNY(converted.OutputPricePerTokenPriority)
	converted.CacheCreationPricePerToken = usdToCNY(converted.CacheCreationPricePerToken)
	converted.CacheCreationPricePerTokenPriority = usdToCNY(converted.CacheCreationPricePerTokenPriority)
	converted.CacheReadPricePerToken = usdToCNY(converted.CacheReadPricePerToken)
	converted.CacheReadPricePerTokenPriority = usdToCNY(converted.CacheReadPricePerTokenPriority)
	converted.CacheCreation5mPrice = usdToCNY(converted.CacheCreation5mPrice)
	converted.CacheCreation1hPrice = usdToCNY(converted.CacheCreation1hPrice)
	converted.ImageOutputPricePerToken = usdToCNY(converted.ImageOutputPricePerToken)
	return &converted
}
