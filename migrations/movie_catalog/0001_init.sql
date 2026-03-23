-- Create schema for movie catalog
create schema if not exists movie_catalog;

-- Movies table
create table if not exists movie_catalog.movies (
  id text primary key,
  name text not null,
  alternative_name text,
  year int,
  duration_minutes int,
  rating_kp numeric(3,1),
  poster_url text,
  description text,
  kp_url text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- Genres table
create table if not exists movie_catalog.genres (
  slug text primary key,
  label text not null
);

-- Movie-Genre junction table
create table if not exists movie_catalog.movie_genres (
  movie_id text not null references movie_catalog.movies(id) on delete cascade,
  genre_slug text not null references movie_catalog.genres(slug) on delete cascade,
  primary key (movie_id, genre_slug)
);

-- Indexes
create index if not exists idx_movies_year
  on movie_catalog.movies(year);

create index if not exists idx_movies_duration
  on movie_catalog.movies(duration_minutes);

create index if not exists idx_movie_genres_genre_movie
  on movie_catalog.movie_genres(genre_slug, movie_id);

create index if not exists idx_movie_genres_movie_genre
  on movie_catalog.movie_genres(movie_id, genre_slug);
