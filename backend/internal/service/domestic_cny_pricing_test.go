//go:build unit

package service

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// 断言人民币绝对金额，防止用汇率或被测常量构造预期而掩盖单位回归。
func TestDomesticCNYDeepSeekUnifiedAndDisplay(t *testing.T) {
	bs := NewBillingService(&config.Config{}, nil)
	resolver := NewModelPricingResolver(nil, bs)
	for _, tc := range []struct {
		model               string
		input, output, read float64
	}{
		{"deepseek-v4-flash", 1.5, 4.5, .05}, {"deepseek-v4-pro", 4.5, 13.5, .15},
	} {
		t.Run(tc.model, func(t *testing.T) {
			card, err := bs.GetModelPricing(tc.model)
			require.NoError(t, err)
			require.InDelta(t, tc.input, card.InputPricePerToken*1e6, 1e-10)
			for _, hour := range []int{0, 1, 4, 6, 10} {
				mult := 1.0
				if hour == 1 || hour == 6 {
					mult = 2
				}
				cost, err := bs.CalculateCostUnified(CostInput{Model: tc.model, Tokens: UsageTokens{InputTokens: 1000000, OutputTokens: 1000000, CacheReadTokens: 1000000}, RateMultiplier: 1, PricingAt: time.Date(2026, 9, 7, hour, 0, 0, 0, time.UTC)})
				require.NoError(t, err)
				require.InDelta(t, (tc.input+tc.output+tc.read)*mult, cost.TotalCost, 1e-10)
			}
			sched, err := bs.ResolveContextPricingSchedule(context.Background(), resolver, ContextPricingScheduleInput{Model: tc.model})
			require.NoError(t, err)
			require.Len(t, sched.Tiers, 1)
			require.InDelta(t, tc.input, *sched.Tiers[0].Input*1e6, 1e-10)
			require.True(t, sched.TimePricing.WeekdaysOnly)
			require.Len(t, sched.TimePricing.Periods, 2)
		})
	}
}

const qwenThinkingCNYFixture = `{
 "qwen-plus":{"currency":"CNY","input_cost_per_token":0.0000008,"output_cost_per_token":0.000002,"output_cost_per_token_thinking":0.000008,
 "cache_creation_input_token_cost":0.000001,"cache_read_input_token_cost":0.00000008,
 "pricing_intervals":[
 {"min_tokens":0,"max_tokens":128000,"input_price":0.0000008,"output_price":0.000002,"thinking_output_price":0.000008},
 {"min_tokens":128000,"max_tokens":256000,"input_price":0.0000024,"output_price":0.000020,"thinking_output_price":0.000024},
 {"min_tokens":256000,"input_price":0.0000048,"output_price":0.000048,"thinking_output_price":0.000064}]}}
`

func TestQwenPlusCNYThinkingUsesSameDiscountAtEachTier(t *testing.T) {
	catalog := newStubPricingServiceFromJSON(t, qwenThinkingCNYFixture)
	// 使用测试折扣而不是运营售价；降低成本不得进入此计算输入。
	for _, discount := range []float64{1, .5, .25} {
		bs, resolver := newTokenCostTestEnv(t, PlatformOpenAI, []ChannelModelPricing{{Platform: PlatformOpenAI, Models: []string{"qwen-plus"}, InputPrice: testPtrFloat64(.8e-6 * discount), OutputPrice: testPtrFloat64(2e-6 * discount)}}, catalog)
		g := &Group{ID: 100, Platform: PlatformOpenAI, LongContextPricingEnabled: true, RateMultiplier: 1}
		for _, tc := range []struct {
			input            int
			normal, thinking float64
		}{{128000, 2, 8}, {128001, 20, 24}, {256000, 20, 24}, {256001, 48, 64}} {
			for _, thinking := range []bool{false, true} {
				cost, err := bs.CalculateTokenCostForRequest(TokenCostRequest{Ctx: context.Background(), Model: "qwen-plus", Group: g, Tokens: UsageTokens{InputTokens: tc.input, OutputTokens: 1000000, ThinkingEnabled: thinking}, RateMultiplier: 1, Resolver: resolver})
				require.NoError(t, err)
				want := tc.normal
				if thinking {
					want = tc.thinking
				}
				require.InDelta(t, want*discount, cost.OutputCost, 1e-9)
			}
		}
		for _, thinking := range []bool{false, true} {
			sched, err := bs.ResolveContextPricingSchedule(context.Background(), resolver, ContextPricingScheduleInput{Model: "qwen-plus", Group: g, ThinkingEnabled: thinking})
			require.NoError(t, err)
			require.Len(t, sched.Tiers, 3)
			wants := []float64{2, 20, 48}
			if thinking {
				wants = []float64{8, 24, 64}
			}
			for i := range wants {
				require.InDelta(t, wants[i]*discount, *sched.Tiers[i].Output*1e6, 1e-8)
			}
		}
	}
}

func TestNativeCNYDirectoryAndChannelFallbackAreNotConverted(t *testing.T) {
	catalog := newStubPricingServiceFromJSON(t, qwenThinkingCNYFixture)
	bs := NewBillingService(&config.Config{}, catalog)
	p, err := bs.GetModelPricing("qwen-plus")
	require.NoError(t, err)
	require.Equal(t, BillingCurrency, p.Currency)
	require.InDelta(t, .8e-6, p.InputPricePerToken, 1e-15)
	require.InDelta(t, 24e-6, *p.Intervals[1].ThinkingOutputPrice, 1e-15)
	// 同一源对象多次读取也不累乘；人民币元数据穿透直接目录展示消费者。
	again, err := bs.GetModelPricing("qwen-plus")
	require.NoError(t, err)
	require.Equal(t, p, again)
	require.InDelta(t, .8e-6, catalog.GetModelPricing("qwen-plus").InputCostPerToken, 1e-15)
}

func TestQwenThinkingExplicitSwitchOverridesEffort(t *testing.T) {
	high := "high"
	for _, tc := range []struct {
		body   string
		effort *string
		want   bool
	}{
		{`{"enable_thinking":true}`, nil, true},
		{`{"enable_thinking":false,"reasoning_effort":"high"}`, &high, false},
		{`{"chat_template_kwargs":{"enable_thinking":true}}`, nil, true},
		{`{}`, nil, false},
		{`{"reasoning_effort":"high"}`, &high, false},
	} {
		got := ApplyThinkingEnabledFallback(tc.effort, []byte(tc.body), "qwen-plus")
		require.Equal(t, tc.want, reasoningEffortEnablesThinking(got))
	}
}

func nativeCNYCatalogForTest(t *testing.T) *PricingService {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)
	return newStubPricingServiceFromJSON(t, string(data))
}

func TestOfficialDomesticCatalogNativeCNYCoverage(t *testing.T) {
	catalog := nativeCNYCatalogForTest(t)
	models := []string{"deepseek-v4-flash", "deepseek-v4-pro", "kimi-k3", "kimi-k2.7-code", "kimi-k2.6", "kimi-k2.5", "qwen3.8-max", "qwen3.8-flash", "qwen3.7-max", "qwen3.7-plus", "qwen3.7-flash", "qwen3.6-max-preview", "qwen3.6-plus", "qwen3.6-flash", "qwen3.5-plus", "qwen3.5-flash", "qwen3.5-397b-a17b", "qwen3.5-27b", "qwen-flash", "qwen-plus", "glm-5.3", "glm-5.3-flash", "glm-5.2", "glm-5.2-fast-preview", "glm-5.1", "glm-5"}
	for _, model := range models {
		p := catalog.GetIdentifiedModelPricing(model)
		require.NotNil(t, p, model)
		require.Equal(t, BillingCurrency, p.Currency, model)
	}
	require.InDelta(t, 1.25e-6, catalog.GetModelPricing("qwen3.8-flash").CacheCreationInputTokenCost, 1e-15, "专属官网价覆盖通用125%估算")
	require.InDelta(t, .1e-6, catalog.GetModelPricing("qwen3.8-flash").CacheReadInputTokenCost, 1e-15)
	require.InDelta(t, .188e-6, catalog.GetModelPricing("qwen-flash").CacheCreationInputTokenCost, 1e-15)
	require.False(t, catalog.GetModelPricing("qwen3.5-27b").SupportsPromptCaching)
	require.False(t, catalog.GetModelPricing("qwen3.5-397b-a17b").SupportsPromptCaching)
	require.InDelta(t, 1.3e-6, catalog.GetModelPricing("glm-5.1").CacheReadInputTokenCost, 1e-15)
}

func TestQwenExplicitCacheReadBillingAndPlazaUseSameDiscount(t *testing.T) {
	catalog := nativeCNYCatalogForTest(t)
	bs, resolver := newTokenCostTestEnv(t, PlatformOpenAI, []ChannelModelPricing{{Platform: PlatformOpenAI, Models: []string{"qwen3.8-max"}, InputPrice: testPtrFloat64(6e-6), OutputPrice: testPtrFloat64(18e-6), CacheReadPrice: testPtrFloat64(.75e-6), CacheWritePrice: testPtrFloat64(7.5e-6)}}, catalog)
	group := &Group{ID: 100, Platform: PlatformOpenAI, LongContextPricingEnabled: true}
	for _, explicit := range []bool{false, true} {
		cost, err := bs.CalculateTokenCostForRequest(TokenCostRequest{Ctx: context.Background(), Model: "qwen3.8-max", Group: group, Resolver: resolver, RateMultiplier: 1, Tokens: UsageTokens{CacheReadTokens: 1000000, ExplicitCache: explicit}})
		require.NoError(t, err)
		want := .75
		if explicit {
			want = .5
		}
		require.InDelta(t, want, cost.TotalCost, 1e-10)
	}
	schedule, err := bs.ResolveContextPricingSchedule(context.Background(), resolver, ContextPricingScheduleInput{Model: "qwen3.8-max", Group: group})
	require.NoError(t, err)
	require.InDelta(t, .75e-6, *schedule.Tiers[0].CacheRead, 1e-15)
	require.InDelta(t, .5e-6, *schedule.Tiers[0].ExplicitCacheRead, 1e-15)
	plaza := &ModelPlazaService{billingService: bs, resolver: resolver}
	model := PlazaModel{Name: "qwen3.8-max", Platform: PlatformOpenAI}
	plaza.fillDisplayPricing(context.Background(), &model, group)
	require.InDelta(t, .5e-6, *model.Pricing.ExplicitCacheReadPrice, 1e-15)
}

func TestRMBReferenceRemainsAuthoritativeOverUSDRemoteMirror(t *testing.T) {
	svc := NewPricingService(&config.Config{Pricing: config.PricingConfig{FallbackFile: filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json")}}, nil)
	data := svc.mergeFallbackPricingData(map[string]*LiteLLMModelPricing{"qwen3.8-flash": {InputCostPerToken: 99, OutputCostPerToken: 99}})
	require.Equal(t, BillingCurrency, data["qwen3.8-flash"].Currency)
	require.InDelta(t, .8e-6, data["qwen3.8-flash"].InputCostPerToken, 1e-15)
}

func TestNativeGLMContextBoundaryUsesPublishedInclusiveHighTier(t *testing.T) {
	bs := NewBillingService(&config.Config{}, nativeCNYCatalogForTest(t))
	for _, n := range []int{31999, 32000} {
		cost, err := bs.CalculateCost("glm-5.1", UsageTokens{InputTokens: n, OutputTokens: 1000000}, 1)
		require.NoError(t, err)
		out := 24.0
		if n == 32000 {
			out = 28
		}
		require.InDelta(t, out, cost.OutputCost, 1e-9)
	}
}

func TestStandaloneBinaryCNYFallbackUsesSameCatalog(t *testing.T) {
	bs := NewBillingService(&config.Config{}, nil)
	for _, model := range []string{"qwen3.8-flash", "kimi-k3", "glm-5.3", "deepseek-v4-flash"} {
		actual, err := bs.GetModelPricing(model)
		require.NoError(t, err)
		require.Equal(t, BillingCurrency, actual.Currency)
	}
	svc := NewPricingService(&config.Config{Pricing: config.PricingConfig{FallbackFile: "./resources/model-pricing/model_prices_and_context_window.json"}}, nil)
	data, err := svc.readPricingFile(svc.cfg.Pricing.FallbackFile)
	require.NoError(t, err)
	parsed, err := svc.parsePricingData(data)
	require.NoError(t, err)
	require.Equal(t, BillingCurrency, parsed["qwen3.8-flash"].Currency)
	_, err = svc.readPricingFile(filepath.Join(t.TempDir(), "missing-custom-pricing.json"))
	require.Error(t, err)
}

func TestUnifiedWithoutResolverRetainsExplicitLongContextOptOut(t *testing.T) {
	bs := NewBillingService(&config.Config{}, nil)
	enabled := false
	cost, err := bs.CalculateCostUnified(CostInput{Model: "glm-5.1", Tokens: UsageTokens{InputTokens: 40000, OutputTokens: 1000000}, RateMultiplier: 1, LongContextBillingEnabled: &enabled})
	require.NoError(t, err)
	require.InDelta(t, 24, cost.OutputCost, 1e-10)
}

func TestContextPricingLabelsPreserveExactTokenBoundaries(t *testing.T) {
	for n, want := range map[int]string{31999: "31.999K", 32000: "32K", 128000: "128K", 1000000: "1M", 1234567: "1.234567M"} {
		require.Equal(t, want, formatContextTokenCount(n))
	}
	bs := NewBillingService(&config.Config{}, nil)
	schedule, err := bs.ResolveContextPricingSchedule(context.Background(), NewModelPricingResolver(nil, bs), ContextPricingScheduleInput{Model: "glm-5.1"})
	require.NoError(t, err)
	require.Len(t, schedule.Tiers, 2)
	require.Equal(t, "≤31.999K", schedule.Tiers[0].Label)
	require.Equal(t, ">31.999K", schedule.Tiers[1].Label)
}

func TestNativeCNYTimezoneWithoutHostZoneData(t *testing.T) {
	if os.Getenv("GO_TEST_EMBEDDED_TIMEZONE") == "1" {
		_, err := loadChannelTimePricingLocation("Asia/Shanghai")
		require.NoError(t, err)
		at := time.Date(2026, 9, 7, 1, 0, 0, 0, time.UTC)
		require.Equal(t, 2.0, deepseekOfficialTimePricing.MultiplierAt(at))
		return
	}
	// 新进程在没有 GOROOT/ZONEINFO 数据的情况下验证发行包自身的时区来源。
	executable, err := os.Executable()
	require.NoError(t, err)
	dir := t.TempDir()
	cmd := exec.Command(executable, "-test.run=^TestNativeCNYTimezoneWithoutHostZoneData$")
	cmd.Env = append(os.Environ(), "GOROOT="+dir, "ZONEINFO="+dir, "GO_TEST_EMBEDDED_TIMEZONE=1")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}
