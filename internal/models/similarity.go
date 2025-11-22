package models

// Neighbor representa una película vecina con su similitud
type Neighbor struct {
	MovieID int     `json:"movieId"`
	IIdx    int     `json:"iIdx"`
	Sim     float64 `json:"sim"`
}

// SimilarityDoc representa el documento de similitudes en MongoDB
type SimilarityDoc struct {
	ID        string     `json:"_id"`
	MovieID   int        `json:"movieId"`
	IIdx      int        `json:"iIdx"`
	Metric    string     `json:"metric"`
	K         int        `json:"k"`
	Neighbors []Neighbor `json:"neighbors"`
	UpdatedAt string     `json:"updatedAt"`
}
