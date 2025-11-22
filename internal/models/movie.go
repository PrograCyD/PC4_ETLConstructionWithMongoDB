package models

// Links contiene los enlaces externos de una película
type Links struct {
	Movielens string `json:"movielens,omitempty"`
	IMDB      string `json:"imdb,omitempty"`
	TMDB      string `json:"tmdb,omitempty"`
}

// GenomeTag representa un tag del genome con su relevancia
type GenomeTag struct {
	Tag       string  `json:"tag"`
	Relevance float64 `json:"relevance"`
}

// CastMember representa un miembro del elenco
type CastMember struct {
	Name       string `json:"name"`
	ProfileURL string `json:"profileUrl,omitempty"`
}

// ExternalData contiene datos externos obtenidos de TMDB
type ExternalData struct {
	PosterURL   string       `json:"posterUrl,omitempty"`
	Overview    string       `json:"overview,omitempty"`
	Cast        []CastMember `json:"cast,omitempty"`
	Director    string       `json:"director,omitempty"`
	Runtime     int          `json:"runtime,omitempty"`
	Budget      int          `json:"budget,omitempty"`
	Revenue     int64        `json:"revenue,omitempty"`
	TMDBFetched bool         `json:"tmdbFetched"`
}

// MovieDoc representa el documento completo de una película en MongoDB
type MovieDoc struct {
	MovieID      int           `json:"movieId"`
	IIdx         *int          `json:"iIdx,omitempty"`
	Title        string        `json:"title"`
	Year         *int          `json:"year,omitempty"`
	Genres       []string      `json:"genres"`
	Links        *Links        `json:"links,omitempty"`
	GenomeTags   []GenomeTag   `json:"genomeTags,omitempty"`
	UserTags     []string      `json:"userTags,omitempty"`
	RatingStats  *RatingStats  `json:"ratingStats,omitempty"`
	ExternalData *ExternalData `json:"externalData,omitempty"`
	CreatedAt    string        `json:"createdAt"`
	UpdatedAt    string        `json:"updatedAt"`
}

// TMDBMovieResponse representa la respuesta de la API de TMDB para detalles de película
type TMDBMovieResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	PosterPath  string `json:"poster_path"`
	Runtime     int    `json:"runtime"`
	Budget      int    `json:"budget"`
	Revenue     int64  `json:"revenue"`
	ReleaseDate string `json:"release_date"`
}

// TMDBCreditsResponse representa la respuesta de la API de TMDB para créditos
type TMDBCreditsResponse struct {
	Cast []struct {
		Name        string `json:"name"`
		Character   string `json:"character"`
		Order       int    `json:"order"`
		ProfilePath string `json:"profile_path"`
	} `json:"cast"`
	Crew []struct {
		Name string `json:"name"`
		Job  string `json:"job"`
	} `json:"crew"`
}
