package service

// 国产供应商 Anthropic 协议转换路径的上游读间隔超时回归测试（B3）：
// 上游挂住 SSE（不发数据也不断连）时，CC×anthropic / Responses×anthropic
// 的读循环必须按 gateway.stream_data_interval_timeout 结束，而不是永久阻塞。

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
)

type nativeAnthropicTailErrorReader struct {
	data []byte
	off  int
}

func (r *nativeAnthropicTailErrorReader) Read(p []byte) (int, error) {
	if r.off < len(r.data) {
		n := copy(p, r.data[r.off:])
		r.off += n
		return n, nil
	}
	return 0, errors.New("simulated upstream read failure")
}

func (r *nativeAnthropicTailErrorReader) Close() error { return nil }

func newNativeAnthropicHangTestService(intervalSec int) *OpenAIGatewayService {
	return &OpenAIGatewayService{
		cfg: &config.Config{
			Gateway: config.GatewayConfig{
				StreamDataIntervalTimeout: intervalSec,
				MaxLineSize:               defaultMaxLineSize,
			},
		},
	}
}

func newHangingUpstreamResponse() (*http.Response, *io.PipeReader, *io.PipeWriter) {
	pr, pw := io.Pipe()
	return &http.Response{StatusCode: http.StatusOK, Body: pr, Header: http.Header{}}, pr, pw
}

// miniAnthropicSSEStream 是一段最小可转换的 Anthropic 事件流。
func miniAnthropicSSEStream() string {
	return strings.Join([]string{
		"event: message_start",
		`data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"glm-4.7","usage":{"input_tokens":10,"output_tokens":1}}}`,
		"",
		"event: content_block_start",
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
		"",
		"event: content_block_stop",
		`data: {"type":"content_block_stop","index":0}`,
		"",
		"event: message_delta",
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}`,
		"",
		"event: message_stop",
		`data: {"type":"message_stop"}`,
		"",
		"",
	}, "\n")
}

func toolAnthropicSSEStream() string {
	return strings.Join([]string{
		"event: message_start",
		`data: {"type":"message_start","message":{"id":"msg_tool","type":"message","role":"assistant","content":[],"model":"glm-4.7","usage":{"input_tokens":10}}}`,
		"",
		"event: content_block_start",
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"lookup","input":{}}}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"query\":\"status\"}"}}`,
		"",
		"event: content_block_stop",
		`data: {"type":"content_block_stop","index":0}`,
		"",
		"event: message_delta",
		`data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":5}}`,
		"",
		"event: message_stop",
		`data: {"type":"message_stop"}`,
		"",
	}, "\n")
}

func incompleteAnthropicSSEStream(ending string) string {
	lines := []string{
		"event: message_start",
		`data: {"type":"message_start","message":{"id":"msg_partial","type":"message","role":"assistant","content":[],"model":"glm-4.7","usage":{"input_tokens":10,"output_tokens":1}}}`,
		"",
		"event: content_block_start",
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"partial"}}`,
		"",
		"event: message_delta",
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}`,
		"",
	}
	if ending == "event_error" {
		lines = append(lines,
			"event: error",
			`data: {"type":"error","error":{"type":"overloaded_error","message":"overloaded"}}`,
			"",
		)
	}
	return strings.Join(lines, "\n")
}

func incompleteAnthropicResponse(mode string) *http.Response {
	body := []byte(incompleteAnthropicSSEStream(mode))
	reader := io.NopCloser(strings.NewReader(string(body)))
	if mode == "read_error" {
		reader = &nativeAnthropicTailErrorReader{data: body}
	}
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: reader}
}

func TestAnthropicNativeLinePump_TimesOutWithoutData(t *testing.T) {
	pr, _ := io.Pipe()
	scanner := bufio.NewScanner(pr)
	defer func() { _ = pr.Close() }()

	pump := newAnthropicNativeLinePump(scanner, 50*time.Millisecond)
	defer pump.stop()

	start := time.Now()
	_, err := pump.next()
	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") {
		t.Fatalf("expected interval timeout, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("timeout not respected: %v", elapsed)
	}
}

func TestAnthropicNativeLinePump_DataResetsTimer(t *testing.T) {
	pr, pw := io.Pipe()
	scanner := bufio.NewScanner(pr)
	pump := newAnthropicNativeLinePump(scanner, 1*time.Second)
	defer pump.stop()

	go func() {
		_, _ = pw.Write([]byte("event: ping\n"))
		// 保持流打开且不再发数据：第二次 next 必须超时。
		time.Sleep(3 * time.Second)
		_ = pw.Close()
	}()
	defer func() { _ = pr.Close() }()

	line, err := pump.next()
	if err != nil || line != "event: ping" {
		t.Fatalf("expected first line, got %q err=%v", line, err)
	}

	start := time.Now()
	_, err = pump.next()
	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") {
		t.Fatalf("expected interval timeout after data stops, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("timeout not respected: %v", elapsed)
	}
}

func TestCCStreamingFromNativeAnthropic_HangTimesOut(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(1)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	resp, pr, pw := newHangingUpstreamResponse()
	start := time.Now()
	res, err := svc.handleCCStreamingFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, start, true)
	_ = pw.Close()
	_ = pr.Close()

	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") {
		t.Fatalf("expected stream timeout error, got %v", err)
	}
	if res == nil {
		t.Fatalf("expected result carrying accumulated usage")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("handler did not respect interval bound: %v", elapsed)
	}
}

func TestCCBufferedFromNativeAnthropic_HangTimesOut(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(1)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	resp, pr, pw := newHangingUpstreamResponse()
	start := time.Now()
	_, err := svc.handleCCBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, start)
	_ = pw.Close()
	_ = pr.Close()

	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") {
		t.Fatalf("expected stream timeout error, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("handler did not respect interval bound: %v", elapsed)
	}
	if !strings.Contains(rec.Body.String(), "Upstream stream data interval timeout") {
		t.Fatalf("expected 502 error body, got %q", rec.Body.String())
	}
}

func TestResponsesStreamingFromNativeAnthropic_HangTimesOut(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(1)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	resp, pr, pw := newHangingUpstreamResponse()
	start := time.Now()
	res, err := svc.handleResponsesStreamingFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, start, apicompat.ResponsesClientToolMapping{})
	_ = pw.Close()
	_ = pr.Close()

	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") {
		t.Fatalf("expected stream timeout error, got %v", err)
	}
	if res == nil {
		t.Fatalf("expected result carrying accumulated usage")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("handler did not respect interval bound: %v", elapsed)
	}
}

func TestCCStreamingFromNativeAnthropic_HappyPathStillConverts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(5)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	resp, pr, pw := newHangingUpstreamResponse()
	go func() {
		_, _ = pw.Write([]byte(miniAnthropicSSEStream()))
		_ = pw.Close()
	}()
	defer func() { _ = pr.Close() }()

	res, err := svc.handleCCStreamingFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected result")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Hello") {
		t.Fatalf("expected converted text chunk, got %q", body)
	}
	if !strings.Contains(body, "data: [DONE]") {
		t.Fatalf("expected [DONE] terminator, got %q", body)
	}
}

func TestCCBufferedFromNativeAnthropic_HappyPathStillConverts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(5)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	resp, pr, pw := newHangingUpstreamResponse()
	go func() {
		_, _ = pw.Write([]byte(miniAnthropicSSEStream()))
		_ = pw.Close()
	}()
	defer func() { _ = pr.Close() }()

	res, err := svc.handleCCBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected result")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Hello") {
		t.Fatalf("expected converted text in buffered response, got %q", body)
	}
	if res.Usage.InputTokens != 10 || res.Usage.OutputTokens != 5 {
		t.Fatalf("expected usage 10/5, got %+v", res.Usage)
	}
}

func TestResponsesStreamingFromNativeAnthropic_HappyPathStillConverts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(miniAnthropicSSEStream())),
	}

	result, err := newNativeAnthropicHangTestService(5).handleResponsesStreamingFromNativeAnthropic(
		resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), apicompat.ResponsesClientToolMapping{},
	)
	if err != nil || result == nil {
		t.Fatalf("expected completed result, result=%+v err=%v", result, err)
	}
	if body := rec.Body.String(); !strings.Contains(body, "response.completed") || !strings.Contains(body, "Hello") {
		t.Fatalf("expected normal terminal conversion, got %s", body)
	}
}

func TestCCBufferedFromNativeAnthropic_ToolArgumentsAreValidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(5)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(toolAnthropicSSEStream())), Header: http.Header{}}

	_, err := svc.handleCCBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var body struct {
		Choices []struct {
			Message struct {
				ToolCalls []struct {
					Function struct {
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Choices) != 1 || len(body.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("expected one tool call, got %s", rec.Body.String())
	}
	args := body.Choices[0].Message.ToolCalls[0].Function.Arguments
	if !json.Valid([]byte(args)) || args != `{"query":"status"}` {
		t.Fatalf("expected valid tool arguments, got %q", args)
	}
}

func TestResponsesBufferedFromNativeAnthropic_ToolArgumentsAreValidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(5)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(toolAnthropicSSEStream())), Header: http.Header{}}

	_, err := svc.handleResponsesBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), apicompat.ResponsesClientToolMapping{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var body struct {
		Output []struct {
			Type      string `json:"type"`
			Arguments string `json:"arguments"`
		} `json:"output"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Output) != 1 || body.Output[0].Type != "function_call" {
		t.Fatalf("expected one function call, got %s", rec.Body.String())
	}
	if args := body.Output[0].Arguments; !json.Valid([]byte(args)) || args != `{"query":"status"}` {
		t.Fatalf("expected valid tool arguments, got %q", args)
	}
}

func TestCCNativeAnthropicRejectsIncompleteTerminalStates(t *testing.T) {
	for _, mode := range []string{"eof", "read_error", "event_error"} {
		t.Run("buffered_"+mode, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			result, err := newNativeAnthropicHangTestService(5).handleCCBufferedFromNativeAnthropic(
				incompleteAnthropicResponse(mode), c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(),
			)
			if err == nil || result == nil {
				t.Fatalf("expected partial result and error, result=%+v err=%v", result, err)
			}
			if result.Usage.InputTokens != 10 || result.Usage.OutputTokens != 5 {
				t.Fatalf("expected preserved usage 10/5, got %+v", result.Usage)
			}
			if rec.Code != http.StatusBadGateway {
				t.Fatalf("expected 502 instead of fabricated success, got %d body=%s", rec.Code, rec.Body.String())
			}
			if mode == "event_error" {
				var typed *sseStreamErrorEventError
				if !errors.As(err, &typed) {
					t.Fatalf("expected typed event error, got %T %v", err, err)
				}
			}
		})

		t.Run("streaming_"+mode, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			result, err := newNativeAnthropicHangTestService(5).handleCCStreamingFromNativeAnthropic(
				incompleteAnthropicResponse(mode), c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), true,
			)
			if err == nil || result == nil {
				t.Fatalf("expected partial result and error, result=%+v err=%v", result, err)
			}
			if result.Usage.InputTokens != 10 || result.Usage.OutputTokens != 5 {
				t.Fatalf("expected preserved usage 10/5, got %+v", result.Usage)
			}
			body := rec.Body.String()
			if strings.Contains(body, "data: [DONE]") || strings.Contains(body, `"finish_reason":"stop"`) {
				t.Fatalf("incomplete stream fabricated terminal output: %s", body)
			}
			if mode == "event_error" {
				var typed *sseStreamErrorEventError
				if !errors.As(err, &typed) {
					t.Fatalf("expected typed event error, got %T %v", err, err)
				}
			}
		})
	}
}

func TestResponsesNativeAnthropicRejectsIncompleteTerminalStates(t *testing.T) {
	for _, mode := range []string{"eof", "read_error", "event_error"} {
		t.Run("buffered_"+mode, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			result, err := newNativeAnthropicHangTestService(5).handleResponsesBufferedFromNativeAnthropic(
				incompleteAnthropicResponse(mode), c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), apicompat.ResponsesClientToolMapping{},
			)
			if err == nil || result == nil {
				t.Fatalf("expected partial result and error, result=%+v err=%v", result, err)
			}
			if result.Usage.InputTokens != 10 || result.Usage.OutputTokens != 5 {
				t.Fatalf("expected preserved usage 10/5, got %+v", result.Usage)
			}
			if rec.Code != http.StatusBadGateway {
				t.Fatalf("expected 502 instead of fabricated success, got %d body=%s", rec.Code, rec.Body.String())
			}
			if mode == "event_error" {
				var typed *sseStreamErrorEventError
				if !errors.As(err, &typed) {
					t.Fatalf("expected typed event error, got %T %v", err, err)
				}
			}
		})

		t.Run("streaming_"+mode, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			result, err := newNativeAnthropicHangTestService(5).handleResponsesStreamingFromNativeAnthropic(
				incompleteAnthropicResponse(mode), c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), apicompat.ResponsesClientToolMapping{},
			)
			if err == nil || result == nil {
				t.Fatalf("expected partial result and error, result=%+v err=%v", result, err)
			}
			if result.Usage.InputTokens != 10 || result.Usage.OutputTokens != 5 {
				t.Fatalf("expected preserved usage 10/5, got %+v", result.Usage)
			}
			if body := rec.Body.String(); strings.Contains(body, "response.completed") {
				t.Fatalf("incomplete stream fabricated terminal output: %s", body)
			}
			if mode == "event_error" {
				var typed *sseStreamErrorEventError
				if !errors.As(err, &typed) {
					t.Fatalf("expected typed event error, got %T %v", err, err)
				}
			}
		})
	}
}

func TestBufferedNativeAnthropicIgnoresNegativeContentIndex(t *testing.T) {
	body := strings.Join([]string{
		"event: message_start",
		`data: {"type":"message_start","message":{"id":"msg_negative","type":"message","role":"assistant","content":[],"model":"glm-4.7","usage":{"input_tokens":2,"output_tokens":1}}}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","index":-1,"delta":{"type":"text_delta","text":"bad"}}`,
		"",
		"event: message_stop",
		`data: {"type":"message_stop"}`,
		"",
	}, "\n")

	t.Run("chat", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}
		result, err := newNativeAnthropicHangTestService(5).handleCCBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
		if err != nil || result == nil {
			t.Fatalf("negative index should be ignored, result=%+v err=%v", result, err)
		}
	})

	t.Run("responses", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}
		result, err := newNativeAnthropicHangTestService(5).handleResponsesBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), apicompat.ResponsesClientToolMapping{})
		if err != nil || result == nil {
			t.Fatalf("negative index should be ignored, result=%+v err=%v", result, err)
		}
	})
}
