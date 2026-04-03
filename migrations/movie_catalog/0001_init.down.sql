-- Down migration for movie_catalog/0001_init.sql
BEGIN;

DROP INDEX IF EXISTS movie_catalog.idx_movie_genres_movie_genre;
DROP INDEX IF EXISTS movie_catalog.idx_movie_genres_genre_movie;
DROP INDEX IF EXISTS movie_catalog.idx_movies_duration;
DROP INDEX IF EXISTS movie_catalog.idx_movies_year;
DROP TABLE IF EXISTS movie_catalog.movie_genres;
DROP TABLE IF EXISTS movie_catalog.genres;
DROP TABLE IF EXISTS movie_catalog.movies;
DROP SCHEMA IF EXISTS movie_catalog;

COMMIT;
