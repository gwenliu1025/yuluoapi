package service

import (
	"context"
	"time"
)

// TokenCostRequest 通用网关 token 计费请求。
type TokenCostRequest struct {
	Ctx            context.Context
	Model          string
	Group          *Group
	Tokens         UsageTokens
	RateMultiplier float64
	PricingAt      time.Time
	ServiceTier    string
	Resolver       *ModelPricingResolver
	// Resolved 为调用方预先解析的定价（Resolver.Resolve 的结果），nil 表示未解析。
	Resolved        *ResolvedPricing
	skipTimePricing bool
}

// CalculateTokenCostForRequest 让网关与模型广场共用区间、模式和分时计费入口。
func (s *BillingService) CalculateTokenCostForRequest(req TokenCostRequest) (*CostBreakdown, error) {
	input := s.tokenCostInput(req, req.Resolved)
	if input.Resolver == nil {
		// 目录直算也沿用 token 模式返回契约，而非底层无解析器的兼容空值。
		input.Resolver = NewModelPricingResolver(nil, s)
	}
	return s.CalculateCostUnified(input)
}

func (s *BillingService) tokenCostInput(req TokenCostRequest, resolved *ResolvedPricing) CostInput {
	input := CostInput{
		Ctx:             req.Ctx,
		Model:           req.Model,
		Group:           req.Group,
		Tokens:          req.Tokens,
		RequestCount:    1,
		RateMultiplier:  req.RateMultiplier,
		PricingAt:       req.PricingAt,
		ServiceTier:     req.ServiceTier,
		Resolver:        req.Resolver,
		Resolved:        resolved,
		skipTimePricing: req.skipTimePricing,
	}
	if req.Group != nil {
		gid := req.Group.ID
		input.GroupID = &gid
	}
	return input
}
