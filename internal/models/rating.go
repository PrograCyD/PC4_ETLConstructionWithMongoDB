package models

// RatingStats contiene estadísticas agregadas de ratings para una película
type RatingStats struct {
	Average     float64 `json:"average"`
	Count       int     `json:"count"`
	LastRatedAt string  `json:"lastRatedAt,omitempty"`
}

// RatingDoc representa un rating individual en MongoDB
type RatingDoc struct {
	UserID    int     `json:"userId"`
	MovieID   int     `json:"movieId"`
	Rating    float64 `json:"rating"`
	Timestamp int64   `json:"timestamp"`
}
