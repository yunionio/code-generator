package report

type ResourceKind struct {
	Name string `json:"name"`
}

type Resource struct {
	// Kinds record resource kind
	Kinds         []string `json:"kinds"`
	Keyword       string   `json:"keyword"`
	KeywordPlural string   `json:"keyword_plural"`
	Service       string   `json:"service"`
}
