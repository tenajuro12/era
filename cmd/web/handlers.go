// cmd/web/handlers.go

package main

import (
	"Movies/pkg/forms"
	"Movies/pkg/models"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	s, err := app.movies.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Movies2: s,
	})
}

func (app *application) showMovies(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	s, err := app.movies.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		Movies: s,
	})
}

func (app *application) genre(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	genre := segments[len(segments)-1]

	s, err := app.movies.GetMovieByGenre(genre)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, r, "home.page.tmpl", &templateData{Movies2: s})
}

func (app *application) createMoviesForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{
		Form: forms.New(nil),
	})

}

func (app *application) createMovies(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title", "genre", "released_year", "released_status", "director")
	form.MaxLength("title", 100)
	form.PermittedValues("released_status", "TRUE", "FALSE")

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	rating, err := strconv.ParseFloat(form.Get("rating"), 64)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	releasedYear, err := time.Parse("2006-01-02T15:04", form.Get("released_year"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	released_status, err := strconv.ParseBool(form.Get("released_status"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	id, err := app.movies.Insert(form.Get("title"), form.Get("original_title"), form.Get("genre"), releasedYear, released_status, form.Get("synopsis"),
		rating, form.Get("director"), form.Get("cast"), form.Get("distributor"))
	if errors.Is(err, models.ErrDuplicateMovie) {
		app.clientError(w, http.StatusBadRequest)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Movie successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/movies/%d", id), http.StatusSeeOther)
}

func (app *application) updateMovies(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.serverError(w, err)
		return
	}

	id, err := strconv.Atoi(r.PostForm.Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	title := r.PostForm.Get("title")
	original_title := r.PostForm.Get("original_title")
	genre := r.PostForm.Get("genre")
	synopsis := r.PostForm.Get("synopsis")
	director := r.PostForm.Get("director")
	cast := r.PostForm.Get("cast")
	distributor := r.PostForm.Get("distributor")

	rating, err := strconv.ParseFloat(r.PostForm.Get("released_year"), 64)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	released_year, err := time.Parse("2006-01-02T15:04", r.PostForm.Get("released_year"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	released_status, err := strconv.ParseBool(r.PostForm.Get("released_status"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.movies.Update(title, original_title, genre, released_year, released_status, synopsis, rating, director, cast, distributor)
	if err != nil {
		app.serverError(w, err)
		return
	}

}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MaxLength("name", 255)
	form.MaxLength("email", 255)
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)

	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	// Try to create a new user record in the database. If the email already exists
	// add an error message to the form and re-display it.
	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("role"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.Errors.Add("email", "Address is already in use")
			app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Otherwise add a confirmation flash message to the session confirming that
	// their signup worked and asking them to log in.
	app.session.Put(r, "flash", "Your signup was successful. Please log in.")

	// And redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic error
	// message to the form failures map and re-display the login page.
	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("generic", "Email or Password is incorrect")
			app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Add the ID of the current user to the session, so that they are now 'logged
	// in'.
	app.session.Put(r, "authenticatedUserID", id)

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/movies/create", http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	// Remove the authenticatedUserID from the session data so that the user is
	// 'logged out'.
	app.session.Remove(r, "authenticatedUserID")
	// Add a flash message to the session to confirm to the user that they've been
	// logged out.
	app.session.Put(r, "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userProfile(w http.ResponseWriter, r *http.Request) {
	userID := app.session.GetInt(r, "authenticatedUserID")

	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "profile.page.tmpl", &templateData{
		User: user,
	})
}
