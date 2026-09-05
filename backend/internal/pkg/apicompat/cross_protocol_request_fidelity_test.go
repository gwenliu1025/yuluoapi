package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestChatToAnthropicPreservesCacheControlAndThinkingSwitch(t *testing.T) {
	var req ChatCompletionsRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"model":"qwen-plus",
		"chat_template_kwargs":{"enable_thinking":true},
		"messages":[
			{"role":"system","content":[{"type":"text","text":"stable","cache_control":{"type":"ephemeral"}}]},
			{"role":"user","content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]},
			{"role":"assistant","content":[{"type":"text","text":"prior","cache_control":{"type":"ephemeral"}}]}
		]
	}`), &req))

	responsesReq, err := ChatCompletionsToResponses(&req)
	require.NoError(t, err)
	require.NotNil(t, responsesReq.EnableThinking)
	require.True(t, *responsesReq.EnableThinking)

	anthropicReq, err := ResponsesToAnthropicRequest(responsesReq)
	require.NoError(t, err)
	require.NotNil(t, anthropicReq.Thinking)
	require.Equal(t, "enabled", anthropicReq.Thinking.Type)
	require.Equal(t, "ephemeral", jsonPathString(anthropicReq.System, "0.cache_control.type"))
	require.Equal(t, "ephemeral", jsonPathString(anthropicReq.Messages[0].Content, "0.cache_control.type"))
	require.Equal(t, "ephemeral", jsonPathString(anthropicReq.Messages[1].Content, "0.cache_control.type"))
}

func TestChatToAnthropicTopLevelThinkingFalseWins(t *testing.T) {
	var req ChatCompletionsRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"model":"qwen-plus",
		"enable_thinking":false,
		"chat_template_kwargs":{"enable_thinking":true},
		"reasoning_effort":"high",
		"messages":[{"role":"user","content":"hello"}]
	}`), &req))

	responsesReq, err := ChatCompletionsToResponses(&req)
	require.NoError(t, err)
	anthropicReq, err := ResponsesToAnthropicRequest(responsesReq)
	require.NoError(t, err)
	require.NotNil(t, anthropicReq.Thinking)
	require.Equal(t, "disabled", anthropicReq.Thinking.Type)
}

func TestResponsesToChatPreservesCacheControlAndThinkingSwitch(t *testing.T) {
	var req ResponsesRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"model":"qwen-plus",
		"enable_thinking":true,
		"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hello","cache_control":{"type":"ephemeral"}}]}]
	}`), &req))

	chatReq, err := ResponsesToChatCompletionsRequest(&req)
	require.NoError(t, err)
	require.NotNil(t, chatReq.EnableThinking)
	require.True(t, *chatReq.EnableThinking)
	require.Equal(t, "ephemeral", jsonPathString(chatReq.Messages[0].Content, "0.cache_control.type"))
}

func TestResponsesToChatPreservesCachedTextShapesForEveryMessageRole(t *testing.T) {
	var req ResponsesRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"model":"qwen3.8-max",
		"input":[
			{"type":"message","role":"system","content":[{"type":"input_text","text":"stable","cache_control":{"type":"ephemeral"}}]},
			{"type":"message","role":"assistant","content":[{"type":"output_text","text":"prior","cache_control":{"type":"ephemeral"}}]},
			{"type":"message","role":"user","content":{"type":"input_text","text":"hello","cache_control":{"type":"ephemeral"}}}
		]
	}`), &req))

	chatReq, err := ResponsesToChatCompletionsRequest(&req)
	require.NoError(t, err)
	require.Len(t, chatReq.Messages, 3)
	for _, message := range chatReq.Messages {
		require.Equal(t, "ephemeral", jsonPathString(message.Content, "0.cache_control.type"))
	}
}

func TestAnthropicBridgesPreserveCacheControl(t *testing.T) {
	var req AnthropicRequest
	require.NoError(t, json.Unmarshal([]byte(`{
		"model":"qwen-plus",
		"max_tokens":32,
		"thinking":{"type":"enabled","budget_tokens":1024},
		"system":[{"type":"text","text":"stable","cache_control":{"type":"ephemeral"}}],
		"messages":[
			{"role":"user","content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]},
			{"role":"assistant","content":[{"type":"text","text":"prior","cache_control":{"type":"ephemeral"}}]}
		]
	}`), &req))

	responsesReq, err := AnthropicToResponses(&req)
	require.NoError(t, err)
	require.Equal(t, "ephemeral", jsonPathString(responsesReq.Input, "0.content.0.cache_control.type"))
	require.Equal(t, "ephemeral", jsonPathString(responsesReq.Input, "1.content.0.cache_control.type"))
	require.Equal(t, "ephemeral", jsonPathString(responsesReq.Input, "2.content.0.cache_control.type"))

	chatReq, err := AnthropicToChatCompletionsRequest(&req)
	require.NoError(t, err)
	require.Equal(t, "ephemeral", jsonPathString(chatReq.Messages[0].Content, "0.cache_control.type"))
	require.Equal(t, "ephemeral", jsonPathString(chatReq.Messages[1].Content, "0.cache_control.type"))
	require.Equal(t, "ephemeral", jsonPathString(chatReq.Messages[2].Content, "0.cache_control.type"))
}

func jsonPathString(raw json.RawMessage, path string) string {
	return gjson.GetBytes(raw, path).String()
}
