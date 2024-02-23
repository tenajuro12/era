//models/mysql/movies.go

package mysql

import (
	"Movies/pkg/models"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type MoviesModel struct {
	DB *sql.DB
}

func (m *MoviesModel) Get(id int) (*models.Movies, error) {

	stmt := `SELECT id, title, original_title, genre, released_year, released_status, synopsis, rating, director, cast, distributor
    FROM movies
    WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	s := &models.Movies{}

	err := row.Scan(&s.ID, &s.Title, &s.Original_title, &s.Genre, &s.Released_year, &s.Released_status, &s.Synopsis, &s.Rating, &s.Director, &s.Cast, &s.Distributor)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil

}

func (m *MoviesModel) Latest() ([]*models.Movies, error) {
	stmt := `SELECT id, title, original_title, genre, released_year, released_status, synopsis, rating, director, cast, distributor
	FROM movies ORDER BY released_year DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	movies := []*models.Movies{}
	for rows.Next() {
		s := &models.Movies{}

		err := rows.Scan(&s.ID, &s.Title, &s.Original_title, &s.Genre, &s.Released_year, &s.Released_status, &s.Synopsis, &s.Rating, &s.Director, &s.Cast, &s.Distributor)
		if err != nil {
			return nil, err
		}
		movies = append(movies, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m *MoviesModel) Insert(title, originalTitle, genre string, released_year time.Time, released_status bool, synopsis string, rating float64, director, cast, distributor string) (int, error) {
	stmt := `INSERT INTO movies (title, original_title, genre, released_year, released_status, synopsis, rating, director, cast, distributor)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.Exec(stmt, title, originalTitle, genre, released_year, released_status, synopsis, rating, director, cast, distributor)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil || isDuplicateError(err) {
		return 0, models.ErrDuplicateMovie
	}

	return int(id), nil

}

func (m *MoviesModel) GetMovieByGenre(genre string) ([]*models.Movies, error) {
	stmt := `
        SELECT id, title, original_title, genre, released_year, released_status, synopsis, rating, director, cast, distributor
        FROM movies
        WHERE genre = ?
        ORDER BY released_year DESC`

	rows, err := m.DB.Query(stmt, genre)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*models.Movies

	for rows.Next() {
		s := &models.Movies{}
		err := rows.Scan(&s.ID, &s.Title, &s.Original_title, &s.Genre, &s.Released_year, &s.Released_status, &s.Synopsis, &s.Rating, &s.Director, &s.Cast, &s.Distributor)
		if err != nil {
			return nil, err
		}
		movies = append(movies, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func isDuplicateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Error 1062:")
}

func (m *MoviesModel) Update(title, original_title, genre string, released_year time.Time, released_status bool, synopsis string, rating float64, director, cast, distributor string) error {
	stmt := `UPDATE movies SET title=?, original_title=?, genre=?, released_year=?, released_status=?, synopsis=?, 
                rating=?, director=?, cast=?, distributor=?
              WHERE id=?`

	_, err := m.DB.Exec(stmt, title, original_title, genre, released_year, released_status, synopsis, rating, director, cast, distributor)
	if err != nil {
		return err
	}

	return nil
}
