package domain

import (
	"encoding/json"
	"testing"
)

func TestMovieSummary_JSONRoundTrip(t *testing.T) {
	original := MovieSummary{
		ID:              "tt1234567",
		Name:            "Inception",
		AlternativeName: "Начало",
		Year:            intPtr(2010),
		Duration:        intPtr(148),
		Rating:          floatPtr(8.8),
		PosterURL:       "https://example.com/poster.jpg",
		Description:     "A thief who steals corporate secrets...",
		KPURL:           "https://kinopoisk.ru/film/1234567",
		Genres:          []string{"Sci-Fi", "Thriller", "Action"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded MovieSummary
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.AlternativeName != original.AlternativeName {
		t.Errorf("AlternativeName = %q, want %q", decoded.AlternativeName, original.AlternativeName)
	}
	if decoded.Year == nil || *decoded.Year != *original.Year {
		t.Errorf("Year = %v, want %v", decoded.Year, original.Year)
	}
	if decoded.Duration == nil || *decoded.Duration != *original.Duration {
		t.Errorf("Duration = %v, want %v", decoded.Duration, original.Duration)
	}
	if decoded.Rating == nil || *decoded.Rating != *original.Rating {
		t.Errorf("Rating = %v, want %v", decoded.Rating, original.Rating)
	}
	if decoded.PosterURL != original.PosterURL {
		t.Errorf("PosterURL = %q, want %q", decoded.PosterURL, original.PosterURL)
	}
	if decoded.Description != original.Description {
		t.Errorf("Description = %q, want %q", decoded.Description, original.Description)
	}
	if decoded.KPURL != original.KPURL {
		t.Errorf("KPURL = %q, want %q", decoded.KPURL, original.KPURL)
	}
	if len(decoded.Genres) != len(original.Genres) {
		t.Errorf("Genres len = %d, want %d", len(decoded.Genres), len(original.Genres))
	}
	for i := range original.Genres {
		if decoded.Genres[i] != original.Genres[i] {
			t.Errorf("Genres[%d] = %q, want %q", i, decoded.Genres[i], original.Genres[i])
		}
	}
}

func TestMovieSummary_JSONOmitEmpty(t *testing.T) {
	minimal := MovieSummary{
		ID:     "tt0000001",
		Name:   "Minimal Movie",
		Genres: []string{},
	}

	data, err := json.Marshal(minimal)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	// Optional fields should be omitted when zero/empty
	if _, ok := decoded["alternative_name"]; ok {
		t.Errorf("alternative_name should be omitted when empty")
	}
	if _, ok := decoded["year"]; ok {
		t.Errorf("year should be omitted when nil")
	}
	if _, ok := decoded["duration"]; ok {
		t.Errorf("duration should be omitted when nil")
	}
	if _, ok := decoded["rating"]; ok {
		t.Errorf("rating should be omitted when nil")
	}
	if _, ok := decoded["poster_url"]; ok {
		t.Errorf("poster_url should be omitted when empty")
	}
	if _, ok := decoded["description"]; ok {
		t.Errorf("description should be omitted when empty")
	}
	if _, ok := decoded["kp_url"]; ok {
		t.Errorf("kp_url should be omitted when empty")
	}
}

func TestGenreOption_JSON(t *testing.T) {
	opt := GenreOption{
		Value: "action",
		Label: "Action",
	}

	data, err := json.Marshal(opt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded GenreOption
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Value != opt.Value || decoded.Label != opt.Label {
		t.Errorf("decoded = %+v, want %+v", decoded, opt)
	}
}

func TestMovieQuery_ZeroValues(t *testing.T) {
	var q MovieQuery

	if q.YearMin != nil {
		t.Errorf("YearMin zero value = %v, want nil", q.YearMin)
	}
	if q.DurationMin != nil {
		t.Errorf("DurationMin zero value = %v, want nil", q.DurationMin)
	}
	if q.IncludeGenre != "" {
		t.Errorf("IncludeGenre zero value = %q, want empty", q.IncludeGenre)
	}
	if q.ExcludeGenres != nil {
		t.Errorf("ExcludeGenres zero value = %v, want nil", q.ExcludeGenres)
	}
	if q.Page != 0 {
		t.Errorf("Page zero value = %d, want 0", q.Page)
	}
	if q.PageSize != 0 {
		t.Errorf("PageSize zero value = %d, want 0", q.PageSize)
	}
}

func TestMoviePage_JSON(t *testing.T) {
	page := MoviePage{
		Items: []MovieSummary{
			{ID: "tt1", Name: "Movie 1", Genres: []string{"Action"}},
			{ID: "tt2", Name: "Movie 2", Genres: []string{"Comedy"}},
		},
		Total:    100,
		Page:     2,
		PageSize: 20,
	}

	data, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded MoviePage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Total != page.Total {
		t.Errorf("Total = %d, want %d", decoded.Total, page.Total)
	}
	if decoded.Page != page.Page {
		t.Errorf("Page = %d, want %d", decoded.Page, page.Page)
	}
	if decoded.PageSize != page.PageSize {
		t.Errorf("PageSize = %d, want %d", decoded.PageSize, page.PageSize)
	}
	if len(decoded.Items) != len(page.Items) {
		t.Errorf("Items len = %d, want %d", len(decoded.Items), len(page.Items))
	}
}

func intPtr(i int) *int     { return &i }
func floatPtr(f float64) *float64 { return &f }