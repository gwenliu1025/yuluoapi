//go:build unit

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func qwenNativeAnthropicTestAccount() *Account {
	account := nativeAnthropicTestAccount()
	account.Name = "qwen-native"
	return account
}

func responsesSupportedAPIKeyTestAccount() *Account {
	return &Account{
		ID:          703,
		Name:        "responses-api-key",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{"api_key": "sk-test"},
		Extra:       map[string]any{"openai_responses_supported": true},
	}
}

func completedResponsesStream(model string) *http.Response {
	body := strings.Join([]string{
		`data: {"type":"response.output_text.delta","delta":"ok"}`,
		"",
		`data: {"type":"response.completed","response":{"id":"resp_qwen","object":"response","model":"` + model + `","status":"completed","output":[{"type":"message","id":"msg_1","role":"assistant","status":"completed","content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func directResponsesAccount(passthrough bool) *Account {
	account := responsesSupportedAPIKeyTestAccount()
	account.Credentials["model_mapping"] = map[string]any{"channel-qwen": "qwen3.8-max"}
	if passthrough {
		account.Extra["openai_passthrough"] = true
	}
	return account
}

func completedResponsesJSON(model string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"id":"resp_direct","object":"response","model":"` + model + `","status":"completed","output":[],"usage":{"input_tokens":3,"output_tokens":2,"total_tokens":5}}`,
		)),
	}
}

func TestChatToNativeAnthropicUsesFinalWireMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{
		"model":"qwen-plus","stream":false,
		"chat_template_kwargs":{"enable_thinking":true},
		"messages":[{"role":"user","content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]}]
	}`)
	upstream := &httpUpstreamRecorder{resp: nativeAnthropicStreamResponse()}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardAsChatCompletions(context.Background(), adaptiveProtocolTestContext("/v1/chat/completions", body), qwenNativeAnthropicTestAccount(), body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "enabled", gjson.GetBytes(upstream.lastBody, "thinking.type").String())
	require.Equal(t, "ephemeral", gjson.GetBytes(upstream.lastBody, "messages.0.content.0.cache_control.type").String())
	require.True(t, result.ExplicitCache)
	require.NotNil(t, result.ReasoningEffort)
	require.Equal(t, "high", *result.ReasoningEffort)
}

func TestResponsesToNativeAnthropicUsesFinalWireMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{
		"model":"qwen-plus","stream":false,"enable_thinking":true,
		"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hello","cache_control":{"type":"ephemeral"}}]}]
	}`)
	upstream := &httpUpstreamRecorder{resp: nativeAnthropicStreamResponse()}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
	c := adaptiveProtocolTestContext("/v1/responses", body)

	result, err := svc.Forward(context.Background(), c, qwenNativeAnthropicTestAccount(), body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "enabled", gjson.GetBytes(upstream.lastBody, "thinking.type").String())
	require.Equal(t, "ephemeral", gjson.GetBytes(upstream.lastBody, "messages.0.content.0.cache_control.type").String())
	require.True(t, result.ExplicitCache)
	require.NotNil(t, result.ReasoningEffort)
	require.Equal(t, "high", *result.ReasoningEffort)
}

func TestMessagesToChatUsesFinalWireMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{
		"model":"qwen-plus","max_tokens":32,"stream":false,
		"thinking":{"type":"enabled","budget_tokens":1024},
		"system":[{"type":"text","text":"stable","cache_control":{"type":"ephemeral"}}],
		"messages":[
			{"role":"user","content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]},
			{"role":"assistant","content":[{"type":"text","text":"prior","cache_control":{"type":"ephemeral"}}]}
		]
	}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"id":"chatcmpl_qwen","object":"chat.completion","model":"qwen-plus","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`,
		)),
	}}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardAsAnthropic(context.Background(), c, forceChatMessagesFallbackAccount(), body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, gjson.GetBytes(upstream.lastBody, "enable_thinking").Bool())
	for _, path := range []string{"messages.0.content.0.cache_control.type", "messages.1.content.0.cache_control.type", "messages.2.content.0.cache_control.type"} {
		require.Equal(t, "ephemeral", gjson.GetBytes(upstream.lastBody, path).String())
	}
	require.True(t, result.ExplicitCache)
	require.NotNil(t, result.ReasoningEffort)
	require.Equal(t, "high", *result.ReasoningEffort)
}

func TestResponsesToChatExplicitFalseWinsInFinalWireMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{
		"model":"qwen-plus","stream":false,"enable_thinking":false,
		"chat_template_kwargs":{"enable_thinking":true},
		"reasoning":{"effort":"high"},
		"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hello","cache_control":{"type":"ephemeral"}}]}]
	}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"id":"chatcmpl_qwen","object":"chat.completion","model":"qwen-plus","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`,
		)),
	}}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	result, err := svc.Forward(context.Background(), c, forceChatResponsesFallbackAccount(), body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, gjson.GetBytes(upstream.lastBody, "enable_thinking").Exists())
	require.False(t, gjson.GetBytes(upstream.lastBody, "enable_thinking").Bool())
	require.Equal(t, "ephemeral", gjson.GetBytes(upstream.lastBody, "messages.0.content.0.cache_control.type").String())
	require.True(t, result.ExplicitCache)
	require.NotNil(t, result.ReasoningEffort)
	require.Equal(t, "none", *result.ReasoningEffort)
}

func TestChatToResponsesUsesFinalWireMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{
		"model":"qwen-plus","stream":false,
		"chat_template_kwargs":{"enable_thinking":true},
		"messages":[{"role":"user","content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]}]
	}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	upstream := &httpUpstreamRecorder{resp: completedResponsesStream("qwen-plus")}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, responsesSupportedAPIKeyTestAccount(), body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, gjson.GetBytes(upstream.lastBody, "enable_thinking").Bool())
	require.Equal(t, "ephemeral", gjson.GetBytes(upstream.lastBody, "input.0.content.0.cache_control.type").String())
	require.True(t, result.ExplicitCache)
	require.NotNil(t, result.ReasoningEffort)
	require.Equal(t, "high", *result.ReasoningEffort)
}

func TestMessagesToResponsesUsesFinalWireMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{
		"model":"qwen-plus","max_tokens":32,"stream":false,
		"thinking":{"type":"enabled","budget_tokens":1024},
		"messages":[{"role":"user","content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]}]
	}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	upstream := &httpUpstreamRecorder{resp: completedResponsesStream("qwen-plus")}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardAsAnthropic(context.Background(), c, responsesSupportedAPIKeyTestAccount(), body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, gjson.GetBytes(upstream.lastBody, "enable_thinking").Bool())
	require.Equal(t, "ephemeral", gjson.GetBytes(upstream.lastBody, "input.0.content.0.cache_control.type").String())
	require.True(t, result.ExplicitCache)
	require.NotNil(t, result.ReasoningEffort)
	require.Equal(t, "high", *result.ReasoningEffort)
}

func TestForeignResponsesBridgeStripsVendorFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{
		"model":"gpt-5.4","stream":false,"enable_thinking":true,
		"messages":[{"role":"user","content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]}]
	}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	upstream := &httpUpstreamRecorder{resp: completedResponsesStream("gpt-5.4")}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardAsChatCompletions(context.Background(), c, responsesSupportedAPIKeyTestAccount(), body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, gjson.GetBytes(upstream.lastBody, "enable_thinking").Exists())
	require.False(t, gjson.GetBytes(upstream.lastBody, "input.0.content.0.cache_control").Exists())
	require.False(t, result.ExplicitCache)
}

func TestForeignResponsesToChatRestoresTextShape(t *testing.T) {
	body := []byte(`{
		"model":"gpt-5.4","stream":false,
		"input":[
			{"type":"message","role":"system","content":[{"type":"input_text","text":"stable","cache_control":{"type":"ephemeral"}}]},
			{"type":"message","role":"assistant","content":[{"type":"output_text","text":"prior","cache_control":{"type":"ephemeral"}}]},
			{"type":"message","role":"user","content":{"type":"input_text","text":"hello","cache_control":{"type":"ephemeral"}}}
		]
	}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"id":"chatcmpl_gpt","object":"chat.completion","model":"gpt-5.4","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`,
		)),
	}}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	result, err := svc.Forward(context.Background(), c, forceChatResponsesFallbackAccount(), body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "stable", gjson.GetBytes(upstream.lastBody, "messages.0.content").String())
	require.Equal(t, "prior", gjson.GetBytes(upstream.lastBody, "messages.1.content").String())
	require.Equal(t, "hello", gjson.GetBytes(upstream.lastBody, "messages.2.content").String())
	require.False(t, result.ExplicitCache)
}

func TestForeignChatBridgeStripsVendorCacheControl(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{
		"model":"gpt-5.4","max_tokens":32,"stream":false,
		"thinking":{"type":"enabled","budget_tokens":1024},
		"system":[{"type":"text","text":"stable","cache_control":{"type":"ephemeral"}}],
		"messages":[
			{"role":"user","content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]},
			{"role":"assistant","content":[{"type":"text","text":"prior","cache_control":{"type":"ephemeral"}}]}
		]
	}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"id":"chatcmpl_gpt","object":"chat.completion","model":"gpt-5.4","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`,
		)),
	}}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardAsAnthropic(context.Background(), c, forceChatMessagesFallbackAccount(), body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, gjson.GetBytes(upstream.lastBody, "enable_thinking").Exists())
	require.Equal(t, "stable", gjson.GetBytes(upstream.lastBody, "messages.0.content").String())
	require.Equal(t, "hello", gjson.GetBytes(upstream.lastBody, "messages.1.content").String())
	require.Equal(t, "prior", gjson.GetBytes(upstream.lastBody, "messages.2.content").String())
	require.False(t, gjson.GetBytes(upstream.lastBody, "messages.1.content.0.cache_control").Exists())
	require.False(t, result.ExplicitCache)
}

func TestDirectResponsesFinalWireExplicitCache(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, stream := range []bool{false, true} {
			name := fmt.Sprintf("passthrough_%v_stream_%v", passthrough, stream)
			t.Run(name, func(t *testing.T) {
				gin.SetMode(gin.TestMode)
				requestModel := "channel-qwen"
				if passthrough {
					requestModel = "qwen3.8-max"
				}
				body := []byte(fmt.Sprintf(`{
					"model":%q,"stream":%v,
					"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hello","cache_control":{"type":"ephemeral"}}]}]
				}`, requestModel, stream))
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
				upstreamResp := completedResponsesJSON("qwen3.8-max")
				if stream {
					upstreamResp = completedResponsesStream("qwen3.8-max")
				}
				upstream := &httpUpstreamRecorder{resp: upstreamResp}
				svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

				result, err := svc.Forward(context.Background(), c, directResponsesAccount(passthrough), body)
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, "qwen3.8-max", result.UpstreamModel)
				require.True(t, result.ExplicitCache)
				require.Equal(t, "ephemeral", gjson.GetBytes(upstream.lastBody, "input.0.content.0.cache_control.type").String())
			})
		}
	}
}

func TestDirectResponsesPartialResultKeepsExplicitCache(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		for _, serviceTier := range []string{"priority", "flex"} {
			t.Run(fmt.Sprintf("passthrough_%v_%s", passthrough, serviceTier), func(t *testing.T) {
				gin.SetMode(gin.TestMode)
				requestModel := "channel-qwen"
				if passthrough {
					requestModel = "qwen3.8-max"
				}
				body := []byte(fmt.Sprintf(`{
				"model":%q,"stream":true,"service_tier":%q,
				"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hello","cache_control":{"type":"ephemeral"}}]}]
			}`, requestModel, serviceTier))
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
				partial := strings.Join([]string{
					`data: {"id":"resp_partial","object":"chat.completion.chunk","model":"qwen3.8-max","service_tier":"default","choices":[{"index":0,"delta":{"content":"partial"}}]}`,
					"",
				}, "\n")
				upstream := &httpUpstreamRecorder{resp: &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
					Body:       &errTailReader{data: []byte(partial), err: errors.New("simulated upstream read failure")},
				}}
				svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

				result, err := svc.Forward(context.Background(), c, directResponsesAccount(passthrough), body)
				require.Error(t, err)
				require.NotNil(t, result)
				require.True(t, result.ExplicitCache)
				require.Equal(t, "qwen3.8-max", result.UpstreamModel)
				require.Equal(t, "qwen3.8-max", result.UpstreamResponseModel)
				require.False(t, result.UpstreamResponseModelConflict)
				require.NotNil(t, result.ServiceTier)
				require.Equal(t, serviceTier, *result.ServiceTier)
				require.Equal(t, "default", result.UpstreamResponseServiceTier)
			})
		}
	}
}

func TestDirectResponsesSessionCacheIsNotExplicitCache(t *testing.T) {
	body := []byte(`{"model":"gpt-5.4","stream":false,"prompt_cache_key":"session-only","input":"hello"}`)
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprintf("passthrough_%v", passthrough), func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
			account := responsesSupportedAPIKeyTestAccount()
			if passthrough {
				account.Extra["openai_passthrough"] = true
			}
			upstream := &httpUpstreamRecorder{resp: completedResponsesJSON("gpt-5.4")}
			svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

			result, err := svc.Forward(context.Background(), c, account, body)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.False(t, result.ExplicitCache)
		})
	}
}

func TestNativeAnthropicPartialResultKeepsFinalWireMetadata(t *testing.T) {
	tests := []struct {
		name string
		path string
		body []byte
		call func(*OpenAIGatewayService, *gin.Context, *Account, []byte) (*OpenAIForwardResult, error)
	}{
		{
			name: "chat",
			path: "/v1/chat/completions",
			body: []byte(`{"model":"qwen-plus","enable_thinking":true,"stream":false,"messages":[{"role":"user","content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]}]}`),
			call: func(svc *OpenAIGatewayService, c *gin.Context, account *Account, body []byte) (*OpenAIForwardResult, error) {
				return svc.ForwardAsChatCompletions(context.Background(), c, account, body, "", "")
			},
		},
		{
			name: "responses",
			path: "/v1/responses",
			body: []byte(`{"model":"qwen-plus","enable_thinking":true,"stream":false,"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hello","cache_control":{"type":"ephemeral"}}]}]}`),
			call: func(svc *OpenAIGatewayService, c *gin.Context, account *Account, body []byte) (*OpenAIForwardResult, error) {
				return svc.Forward(context.Background(), c, account, body)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: incompleteAnthropicResponse("eof")}
			svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
			c := adaptiveProtocolTestContext(tt.path, tt.body)

			result, err := tt.call(svc, c, qwenNativeAnthropicTestAccount(), tt.body)
			require.Error(t, err)
			require.NotNil(t, result)
			require.True(t, result.ExplicitCache)
			require.NotNil(t, result.ReasoningEffort)
			require.Equal(t, "high", *result.ReasoningEffort)
			require.Equal(t, 10, result.Usage.InputTokens)
			require.Equal(t, 5, result.Usage.OutputTokens)
		})
	}
}
