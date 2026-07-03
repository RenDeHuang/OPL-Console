package api

type Readiness struct {
	Ready  bool            `json:"ready"`
	Checks map[string]bool `json:"checks"`
}
