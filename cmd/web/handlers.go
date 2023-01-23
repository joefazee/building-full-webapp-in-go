package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"abahjoseph.com/snippetbox/pkg/forms"
	"abahjoseph.com/snippetbox/pkg/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	snippets, err := app.snippets.Latest()

	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Snippets: snippets,
	})

}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	s, err := app.snippets.Get(id)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		Snippet: s,
	})
}

func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {

	// If we want to limit the size of the boyd
	r.Body = http.MaxBytesReader(w, r.Body, 1024*2)

	err := r.ParseForm()
	if err != nil {
		app.serverError(w, err)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires").
		MaxLength("title", 100).
		MaxLength("content", 500).
		PermittedValues("expires", "365", "7", "1").
		Custom("title", func(field, value string, obj *forms.Form) {
			if strings.TrimSpace(value) == "aj" {
				obj.Fail(field, "This value is not allowed")
			}
		})

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{
			Form: form,
		})
		return
	}

	id, err := app.snippets.Insert(form.Get("title"), form.Get("content"), form.Get("expires"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Snippet created successfully")
	http.Redirect(w, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther)
}

func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {

	app.render(w, r, "create.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {

	app.render(w, r, "about.page.tmpl", nil)
}

func (app *application) userProfile(w http.ResponseWriter, r *http.Request) {
	userID := app.session.GetInt(r, authenticatedUserID)

	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Fprintf(w, "%+v", user)

}

func (app *application) templateTest(w http.ResponseWriter, r *http.Request) {

	app.render(w, r, "about/aj.page.tmpl", nil)
}
