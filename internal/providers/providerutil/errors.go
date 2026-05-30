package providerutil

import "fmt"

// UnsupportedQuestionError reports an answerable question whose provider type
// has no runtime builder. Silent skips here create malformed submissions.
type UnsupportedQuestionError struct {
	Provider     string
	QuestionNum  int
	TypeCode     string
	ProviderType string
	Reason       string
}

func (e *UnsupportedQuestionError) Error() string {
	reason := e.Reason
	if reason == "" {
		reason = "暂不支持该题型"
	}
	if e.ProviderType != "" {
		return fmt.Sprintf("%s第%d题暂不支持: type_code=%s provider_type=%s, %s", e.Provider, e.QuestionNum, e.TypeCode, e.ProviderType, reason)
	}
	return fmt.Sprintf("%s第%d题暂不支持: type_code=%s, %s", e.Provider, e.QuestionNum, e.TypeCode, reason)
}
