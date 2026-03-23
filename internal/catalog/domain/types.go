package domain

type MovieSummary struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	AlternativeName string   `json:"alternative_name,omitempty"`
	Year            *int     `json:"year,omitempty"`
	Duration        *int     `json:"duration,omitempty"`
	Rating          *float64 `json:"rating,omitempty"`
	PosterURL       string   `json:"poster_url,omitempty"`
	Description     string   `json:"description,omitempty"`
	KPURL           string   `json:"kp_url,omitempty"`
	Genres          []string `json:"genres"`
}

type GenreOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type MovieQuery struct {
	YearMin       *int
	DurationMin   *int
	IncludeGenre  string
	ExcludeGenres []string
	Page          int
	PageSize      int
}

type MoviePage struct {
	Items    []MovieSummary `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
