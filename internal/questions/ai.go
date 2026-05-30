package questions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	aiRequestTimeout = 12 * time.Second
	aiMaxAttempts    = 4
	aiRetryBackoff   = 400 * time.Millisecond
)

// AIConfig holds AI generation configuration.
type AIConfig struct {
	Mode         string // "free", "api"
	Provider     string // "deepseek", etc.
	APIKey       string
	BaseURL      string
	Model        string
	SystemPrompt string
}

// AIClient generates text answers using AI.
type AIClient struct {
	config AIConfig
	client *http.Client
}

// AIError classifies an AI generation failure for callers and tests.
type AIError struct {
	Kind string
	Err  error
}

func (e *AIError) Error() string {
	if e == nil || e.Err == nil {
		return "AI 调用失败"
	}
	return fmt.Sprintf("AI %s: %v", e.Kind, e.Err)
}

func (e *AIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

const (
	AIErrorConfig   = "config"
	AIErrorTimeout  = "timeout"
	AIErrorHTTP     = "http"
	AIErrorResponse = "response"
	AIErrorNetwork  = "network"
)

// NewAIClient creates a new AI client.
func NewAIClient(config AIConfig) *AIClient {
	return &AIClient{
		config: config,
		client: &http.Client{Timeout: aiRequestTimeout},
	}
}

// GenerateAnswer generates a text answer for a question.
func (a *AIClient) GenerateAnswer(questionTitle, questionType string, blankCount int) (string, error) {
	if a.config.Mode == "" || a.config.Mode == "free" {
		return a.generateFree(questionTitle, questionType, blankCount)
	}
	return a.generateAPI(questionTitle, questionType, blankCount)
}

func (a *AIClient) generateFree(questionTitle, questionType string, blankCount int) (string, error) {
	// Free mode uses a simple heuristic response
	if blankCount > 1 {
		answers := make([]string, blankCount)
		defaults := []string{"非常满意", "服务态度好", "环境优美", "效率高", "质量好"}
		for i := 0; i < blankCount && i < len(defaults); i++ {
			answers[i] = defaults[i]
		}
		return strings.Join(answers, "|"), nil
	}
	return "非常满意", nil
}

func (a *AIClient) generateAPI(questionTitle, questionType string, blankCount int) (string, error) {
	if a.config.APIKey == "" {
		return "", classifyAIError(AIErrorConfig, fmt.Errorf("API key 未配置"))
	}

	prompt := a.buildPrompt(questionTitle, questionType, blankCount)
	systemPrompt := a.config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "你是一个问卷答题助手，请根据题目生成合理的答案。只输出答案内容，不要解释。"
	}

	baseURL := a.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	model := a.config.Model
	if model == "" {
		model = "deepseek-chat"
	}

	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  200,
	}

	var lastErr error
	for attempt := 1; attempt <= aiMaxAttempts; attempt++ {
		answer, err := a.doGenerateAPI(baseURL, reqBody)
		if err == nil {
			return answer, nil
		}
		lastErr = err
		if attempt >= aiMaxAttempts || !isRetryableAIError(err) {
			break
		}
		time.Sleep(aiRetryBackoff)
	}
	return "", lastErr
}

func (a *AIClient) doGenerateAPI(baseURL string, reqBody map[string]any) (string, error) {
	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", strings.TrimRight(baseURL, "/")+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", classifyAIError(AIErrorConfig, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return "", classifyAIError(AIErrorTimeout, err)
		}
		return "", classifyAIError(AIErrorNetwork, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		kind := AIErrorHTTP
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			kind = AIErrorConfig
		}
		return "", classifyAIError(kind, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateString(string(respBody), 200)))
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", classifyAIError(AIErrorResponse, err)
	}

	// Extract content from response
	if choices, ok := result["choices"].([]any); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]any); ok {
			if message, ok := choice["message"].(map[string]any); ok {
				if content, ok := message["content"].(string); ok {
					return strings.TrimSpace(content), nil
				}
			}
		}
	}

	return "", classifyAIError(AIErrorResponse, fmt.Errorf("响应格式错误"))
}

func classifyAIError(kind string, err error) error {
	return &AIError{Kind: kind, Err: err}
}

func isRetryableAIError(err error) bool {
	if err == nil {
		return false
	}
	aiErr, ok := err.(*AIError)
	if !ok {
		return true
	}
	switch aiErr.Kind {
	case AIErrorConfig, AIErrorResponse:
		return false
	default:
		return true
	}
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "timeout") || strings.Contains(text, "timed out") || strings.Contains(text, "超时")
}

func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (a *AIClient) buildPrompt(questionTitle, questionType string, blankCount int) string {
	cleaned := cleanQuestionTitle(questionTitle)
	if blankCount > 1 {
		return fmt.Sprintf("题目：%s\n这是一个包含 %d 个空格的填空题，请为每个空格生成一个答案，用 | 分隔。", cleaned, blankCount)
	}
	return fmt.Sprintf("题目：%s\n请生成一个简短的回答。", cleaned)
}

func cleanQuestionTitle(title string) string {
	// Remove numbering
	cleaned := title
	// Remove common prefixes
	for _, prefix := range []string{"Q:", "q:", "问:", "问题:"} {
		cleaned = strings.TrimPrefix(cleaned, prefix)
	}
	return strings.TrimSpace(cleaned)
}
