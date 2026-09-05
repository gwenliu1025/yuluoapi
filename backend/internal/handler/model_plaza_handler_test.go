//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type modelPlazaSettingRepoStub struct {
	service.SettingRepository
}

func (modelPlazaSettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	return map[string]string{
		service.SettingKeyModelPlazaEnabled:     "true",
		service.SettingKeyModelPlazaRequireAuth: "true",
	}, nil
}

type modelPlazaChannelRepoStub struct {
	service.ChannelRepository
	channels []service.Channel
}

func (s modelPlazaChannelRepoStub) ListAll(context.Context) ([]service.Channel, error) {
	return s.channels, nil
}

type modelPlazaGroupRepoStub struct {
	service.GroupRepository
	groups []service.Group
}

func (s modelPlazaGroupRepoStub) ListActive(context.Context) ([]service.Group, error) {
	return s.groups, nil
}

type modelPlazaUserRepoStub struct {
	service.UserRepository
}

func (modelPlazaUserRepoStub) GetByID(_ context.Context, userID int64) (*service.User, error) {
	return &service.User{ID: userID}, nil
}

type modelPlazaSubscriptionRepoStub struct {
	service.UserSubscriptionRepository
}

func (modelPlazaSubscriptionRepoStub) ListActiveByUserID(context.Context, int64) ([]service.UserSubscription, error) {
	return nil, nil
}

type modelPlazaRateRepoStub struct {
	service.UserGroupRateRepository
	err   error
	calls int
}

func (s *modelPlazaRateRepoStub) GetByUserID(context.Context, int64) (map[int64]float64, error) {
	s.calls++
	return nil, s.err
}

func plazaGroups() []service.PlazaGroup {
	return []service.PlazaGroup{
		{ID: 1, Name: "public-standard", Platform: "anthropic", SubscriptionType: "standard", RateMultiplier: 1},
		{ID: 2, Name: "exclusive-a", Platform: "anthropic", IsExclusive: true, RateMultiplier: 0.5},
		{ID: 3, Name: "public-subscription", Platform: "openai", SubscriptionType: "subscription", RateMultiplier: 1},
		{ID: 4, Name: "exclusive-b", Platform: "openai", IsExclusive: true, RateMultiplier: 0.8},
	}
}

func TestFilterPlazaVisibleGroups_AnonymousSeesOnlyNonExclusive(t *testing.T) {
	// 匿名(allowedGroups == nil):仅非专属分组;订阅型公开分组照常可见(橱窗语义)。
	visible := filterPlazaVisibleGroups(plazaGroups(), nil)
	require.Len(t, visible, 2)
	ids := []int64{visible[0].ID, visible[1].ID}
	require.ElementsMatch(t, []int64{1, 3}, ids)
}

func TestFilterPlazaVisibleGroups_AuthedSeesOnlyActuallyAvailableGroups(t *testing.T) {
	// GetAvailableGroups 返回公开标准组 + 获授专属组；无有效订阅的公开订阅组不在集合中。
	allowed := map[int64]struct{}{1: {}, 2: {}}
	visible := filterPlazaVisibleGroups(plazaGroups(), allowed)
	require.Len(t, visible, 2)
	ids := make([]int64, 0, len(visible))
	for _, g := range visible {
		ids = append(ids, g.ID)
	}
	require.ElementsMatch(t, []int64{1, 2}, ids)
}

func TestFilterPlazaVisibleGroups_AuthedEmptySetSeesNothing(t *testing.T) {
	// 空集合仍是登录态的权威结果，不能降级为匿名后重新放行公开组。
	visible := filterPlazaVisibleGroups(plazaGroups(), map[int64]struct{}{})
	require.Empty(t, visible)
}

func TestFilterPlazaVisibleGroups_AuthedSubscriptionRequiresAvailableMembership(t *testing.T) {
	// 有效订阅由 GetAvailableGroups 投影为分组 ID；在集合中时订阅组才可见。
	allowed := map[int64]struct{}{1: {}, 3: {}}
	visible := filterPlazaVisibleGroups(plazaGroups(), allowed)
	ids := make([]int64, 0, len(visible))
	for _, g := range visible {
		ids = append(ids, g.ID)
	}
	require.ElementsMatch(t, []int64{1, 3}, ids)
}

func TestModelPlazaHandler_NilSettingServiceFailsClosed404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ModelPlazaHandler{} // settingService == nil → fail-closed
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/model-plaza", nil)

	h.Get(c)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestModelPlazaHandler_UserRateFailureDoesNotExposeDefaultPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	newRequest := func(rateRepo service.UserGroupRateRepository) *httptest.ResponseRecorder {
		group := service.Group{
			ID: 1, Name: "public-standard", Platform: service.PlatformAnthropic,
			SubscriptionType: "standard", RateMultiplier: 0.8, Status: service.StatusActive,
		}
		groupRepo := modelPlazaGroupRepoStub{groups: []service.Group{group}}
		plazaService := service.NewModelPlazaService(
			modelPlazaChannelRepoStub{channels: []service.Channel{{
				ID: 1, Name: "channel", Status: service.StatusActive, GroupIDs: []int64{1},
				ModelPricing: []service.ChannelModelPricing{{
					Platform: service.PlatformAnthropic, Models: []string{"test-model"},
					BillingMode: service.BillingModeToken, InputPrice: testPtr(1e-6),
				}},
			}}},
			groupRepo,
			nil,
			nil,
			nil,
		)
		apiKeyService := service.NewAPIKeyService(
			nil,
			modelPlazaUserRepoStub{},
			groupRepo,
			modelPlazaSubscriptionRepoStub{},
			rateRepo,
			nil,
			nil,
		)
		h := NewModelPlazaHandler(
			plazaService,
			apiKeyService,
			service.NewSettingService(modelPlazaSettingRepoStub{}, nil),
		)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/model-plaza", nil)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
		h.Get(c)
		return w
	}

	t.Run("查询失败返回错误，不把默认倍率当账号实际价格", func(t *testing.T) {
		rateRepo := &modelPlazaRateRepoStub{err: errors.New("load user group rates")}
		w := newRequest(rateRepo)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Equal(t, 1, rateRepo.calls)
		require.NotContains(t, w.Body.String(), "rate_multiplier")
	})

	t.Run("正常空倍率沿用分组默认倍率", func(t *testing.T) {
		w := newRequest(nil)
		require.Equal(t, http.StatusOK, w.Code)

		var envelope struct {
			Data modelPlazaResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
		require.Len(t, envelope.Data.Groups, 1)
		require.InDelta(t, 0.8, envelope.Data.Groups[0].RateMultiplier, 1e-9)
		require.Nil(t, envelope.Data.Groups[0].UserRateMultiplier)
	})
}

func TestToModelPlazaGroupDTO_UserRateAndFieldWhitelist(t *testing.T) {
	g := service.PlazaGroup{
		ID: 2, Name: "vip", Description: "d", Platform: "anthropic",
		SubscriptionType: "standard", RateMultiplier: 1, IsExclusive: true,
		Models: []service.PlazaModel{{
			Name:     "claude-sonnet",
			Platform: "anthropic",
			Pricing: &service.ChannelModelPricing{
				BillingMode: service.BillingModeToken,
				InputPrice:  testPtr(3e-6),
			},
			OfficialPricing: &service.PlazaOfficialPricing{
				InputPrice:     testPtr(3e-6),
				CacheReadPrice: testPtr(3e-7),
			},
		}},
	}

	// 有专属倍率:user_rate_multiplier 序列化输出
	dto := toModelPlazaGroupDTO(&g, map[int64]float64{2: 0.5})
	raw, err := json.Marshal(dto)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	for _, key := range []string{
		"id", "name", "description", "platform", "subscription_type",
		"rate_multiplier", "user_rate_multiplier", "is_exclusive", "models",
		"peak_rate_enabled", "peak_start", "peak_end", "peak_rate_multiplier",
		"image_rate_independent", "image_rate_multiplier", "long_context_pricing_enabled",
	} {
		_, exists := decoded[key]
		require.Truef(t, exists, "plaza group DTO must expose %q", key)
	}
	require.InDelta(t, 0.5, decoded["user_rate_multiplier"].(float64), 1e-9)

	// 模型条目:pricing + official_pricing 并存;official 缺失字段输出 null 而非省略
	models := decoded["models"].([]any)
	require.Len(t, models, 1)
	model := models[0].(map[string]any)
	require.Contains(t, model, "pricing")
	require.Contains(t, model, "official_pricing")
	official := model["official_pricing"].(map[string]any)
	require.Contains(t, official, "input_price")
	require.Contains(t, official, "cache_read_price")
	_, has1h := official["cache_write_1h_price"]
	require.False(t, has1h, "1h 缓存写价为 nil 时应 omitempty")
	_, hasOfficialIntervals := official["intervals"]
	require.False(t, hasOfficialIntervals, "官方无阶梯时 intervals 应 omitempty")
	_, hasBasis := model["long_context_basis"]
	require.False(t, hasBasis, "单档模型不输出 long_context_basis")
	_, hasTimePricing := model["time_pricing"]
	require.False(t, hasTimePricing, "无分时时不输出 time_pricing")

	// 无专属倍率:user_rate_multiplier 整个字段省略
	dtoNoRate := toModelPlazaGroupDTO(&g, nil)
	rawNoRate, err := json.Marshal(dtoNoRate)
	require.NoError(t, err)
	var decodedNoRate map[string]any
	require.NoError(t, json.Unmarshal(rawNoRate, &decodedNoRate))
	_, hasRate := decodedNoRate["user_rate_multiplier"]
	require.False(t, hasRate, "无专属倍率时 user_rate_multiplier 应 omitempty")
}

func TestToModelPlazaOfficialPricing_NilPassthrough(t *testing.T) {
	require.Nil(t, toModelPlazaOfficialPricing(nil))
}

func TestToModelPlazaGroupDTO_LongContextTiersAndBasis(t *testing.T) {
	maxTokens := 272000
	g := service.PlazaGroup{
		ID: 3, Name: "ladder", Platform: "openai", SubscriptionType: "standard", RateMultiplier: 1,
		LongContextPricingEnabled: true,
		Models: []service.PlazaModel{{
			Name:     "gpt-5.4",
			Platform: "openai",
			Pricing: &service.ChannelModelPricing{
				BillingMode: service.BillingModeToken,
				InputPrice:  testPtr(2.5e-6),
				Intervals: []service.PricingInterval{
					{MinTokens: 0, MaxTokens: &maxTokens, TierLabel: "≤272K", InputPrice: testPtr(2.5e-6)},
					{MinTokens: 272000, TierLabel: ">272K", InputPrice: testPtr(5e-6)},
				},
			},
			OfficialPricing: &service.PlazaOfficialPricing{
				InputPrice: testPtr(2.5e-6),
				Intervals: []service.PricingInterval{
					{MinTokens: 0, MaxTokens: &maxTokens, TierLabel: "≤272K", InputPrice: testPtr(2.5e-6)},
					{MinTokens: 272000, TierLabel: ">272K", InputPrice: testPtr(5e-6)},
				},
			},
			LongContextBasis: service.ContextPricingBasisWholeRequest,
		}},
	}

	raw, err := json.Marshal(toModelPlazaGroupDTO(&g, nil))
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Equal(t, true, decoded["long_context_pricing_enabled"])

	model := decoded["models"].([]any)[0].(map[string]any)
	require.Equal(t, "whole_request", model["long_context_basis"])

	pricing := model["pricing"].(map[string]any)
	paidTiers := pricing["intervals"].([]any)
	require.Len(t, paidTiers, 2)
	require.Equal(t, ">272K", paidTiers[1].(map[string]any)["tier_label"])

	official := model["official_pricing"].(map[string]any)
	officialTiers := official["intervals"].([]any)
	require.Len(t, officialTiers, 2)
	first := officialTiers[0].(map[string]any)
	require.Equal(t, "≤272K", first["tier_label"])
	require.InDelta(t, 272000, first["max_tokens"].(float64), 0)
	require.Contains(t, first, "cache_write_price", "区间 DTO 字段齐全（nil 输出 null）")
}

func testPtr(v float64) *float64 { return &v }

func TestToModelPlazaGroupDTO_TimePricing(t *testing.T) {
	g := service.PlazaGroup{
		ID: 4, Name: "cn", Platform: "deepseek", SubscriptionType: "standard", RateMultiplier: 1,
		Models: []service.PlazaModel{{
			Name:     "deepseek-chat",
			Platform: "deepseek",
			Pricing:  &service.ChannelModelPricing{BillingMode: service.BillingModeToken, InputPrice: testPtr(0.28e-6)},
			TimePricing: &service.TimePricingSchedule{Timezone: "Asia/Shanghai", Periods: []service.TimePricingPeriod{
				{StartTime: "00:30", EndTime: "08:30", Multiplier: 0.5},
			}},
		}, {
			Name:     "deepseek-reasoner",
			Platform: "deepseek",
			Pricing:  &service.ChannelModelPricing{BillingMode: service.BillingModeToken, InputPrice: testPtr(0.56e-6)},
			TimePricing: &service.TimePricingSchedule{Timezone: "Asia/Shanghai", WeekdaysOnly: true, Periods: []service.TimePricingPeriod{
				{StartTime: "00:30", EndTime: "08:30", Multiplier: 0.5},
			}},
		}},
	}
	raw, err := json.Marshal(toModelPlazaGroupDTO(&g, nil))
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	model := decoded["models"].([]any)[0].(map[string]any)
	tp := model["time_pricing"].(map[string]any)
	require.Equal(t, "Asia/Shanghai", tp["timezone"])
	_, hasWeekdaysOnly := tp["weekdays_only"]
	require.False(t, hasWeekdaysOnly, "未开启仅工作日时字段省略")
	periods := tp["periods"].([]any)
	require.Len(t, periods, 1)
	first := periods[0].(map[string]any)
	require.Equal(t, "00:30", first["start_time"])
	require.Equal(t, "08:30", first["end_time"])
	require.InDelta(t, 0.5, first["multiplier"].(float64), 1e-12)

	weekdaysModel := decoded["models"].([]any)[1].(map[string]any)
	weekdaysTP := weekdaysModel["time_pricing"].(map[string]any)
	require.Equal(t, true, weekdaysTP["weekdays_only"])
}
