package model

const ReviewAdviceKind = "advice-v3"

// ReviewAdvice is the AI interpretation stored alongside a published day revision.
type ReviewAdvice struct {
	Market       string `json:"market"`
	Buy          string `json:"buy"`
	AverageState string `json:"averageState"`
	StyleBias    string `json:"styleBias"`
	SectorState  string `json:"sectorState"`
	Model        string `json:"model"`
	GeneratedAt  string `json:"generatedAt"`
	Rule         string `json:"rule"`
}
