//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestAccountShouldIgnoreGemini429RateLimit(t *testing.T) {
	for _, accountType := range []string{AccountTypeAPIKey, AccountTypeOAuth, AccountTypeServiceAccount} {
		t.Run(accountType+"_defaults_disabled", func(t *testing.T) {
			account := &Account{Platform: PlatformGemini, Type: accountType}
			require.False(t, account.ShouldIgnoreGemini429RateLimit())
		})
	}

	tests := []struct {
		name    string
		account *Account
		want    bool
	}{
		{
			name: "explicit_true",
			account: &Account{Platform: PlatformGemini, Extra: map[string]any{
				GeminiIgnore429RateLimitExtraKey: true,
			}},
			want: true,
		},
		{
			name: "explicit_false",
			account: &Account{Platform: PlatformGemini, Extra: map[string]any{
				GeminiIgnore429RateLimitExtraKey: false,
			}},
			want: false,
		},
		{
			name: "malformed_value_disables_policy",
			account: &Account{Platform: PlatformGemini, Extra: map[string]any{
				GeminiIgnore429RateLimitExtraKey: "false",
			}},
			want: false,
		},
		{
			name:    "non_gemini",
			account: &Account{Platform: PlatformOpenAI},
			want:    false,
		},
		{
			name:    "nil_account",
			account: nil,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.account.ShouldIgnoreGemini429RateLimit())
		})
	}
}

func TestGemini429ExplicitPolicySkipsAllLocalSchedulingState(t *testing.T) {
	repo := &geminiErrorPolicyRepo{}
	rateLimitService := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	compatService := &GeminiMessagesCompatService{
		accountRepo:      repo,
		rateLimitService: rateLimitService,
	}
	account := &Account{
		ID:       901,
		Platform: PlatformGemini,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			GeminiIgnore429RateLimitExtraKey: true,
		},
		Credentials: map[string]any{
			"custom_error_codes_enabled": true,
			"custom_error_codes":         []any{float64(http.StatusTooManyRequests)},
			"temp_unschedulable_enabled": true,
			"temp_unschedulable_rules": []any{
				map[string]any{
					"error_code":       float64(http.StatusTooManyRequests),
					"keywords":         []any{"quota"},
					"duration_minutes": float64(10),
				},
			},
		},
	}
	body := []byte(`{"error":{"message":"quota exhausted"}}`)

	// 显式开启后，自定义错误码仍可参与故障转移分类，但实际状态写入必须被账号开关拦截。
	require.Equal(t, ErrorPolicyMatched,
		rateLimitService.CheckErrorPolicy(context.Background(), account, http.StatusTooManyRequests, body))
	require.False(t,
		rateLimitService.HandleUpstreamError(context.Background(), account, http.StatusTooManyRequests, http.Header{}, body))
	require.False(t,
		rateLimitService.HandleTempUnschedulable(context.Background(), account, http.StatusTooManyRequests, body))
	compatService.handleGeminiUpstreamError(context.Background(), account, http.StatusTooManyRequests, http.Header{}, body)

	require.Zero(t, repo.setRateLimitedCalls)
	require.Zero(t, repo.setTempCalls)
	require.Zero(t, repo.setErrorCalls)
	require.Zero(t, repo.setModelRateLimitedCalls)
}
