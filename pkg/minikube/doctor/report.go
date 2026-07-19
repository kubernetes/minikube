package doctor

type Result struct {
	Name           string `json:"name"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	Details        string `json:"details,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
}
