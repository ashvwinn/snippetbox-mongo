package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ASH-WIN-10/snippetbox/internal/models"
	"github.com/ASH-WIN-10/snippetbox/internal/validator"
)

type SnippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type SnippetEditForm struct {
	ID                  string `form:"id"`
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type UserSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type UserLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets

	app.render(w, r, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, r, http.StatusOK, "view.tmpl.html", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = SnippetCreateForm{
		Expires: 365,
	}

	app.render(w, r, http.StatusOK, "create.tmpl.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form SnippetCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl.html", data)
		return
	}

	userId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	snippetId, err := app.snippets.Insert(form.Title, form.Content, form.Expires, userId)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%s", snippetId), http.StatusSeeOther)
}

func (app *application) snippetDeletePost(w http.ResponseWriter, r *http.Request) {
	snippetId := r.PathValue("id")
	if snippetId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	bool, err := app.snippets.CheckSnippetOwnership(snippetId, userId)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !bool {
		app.clientError(w, http.StatusForbidden)
		return
	}

	err = app.snippets.Delete(snippetId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNoRecord):
			http.NotFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet deleted successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) snippetEdit(w http.ResponseWriter, r *http.Request) {
	snippetId := r.PathValue("id")
	if snippetId == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	bool, err := app.snippets.CheckSnippetOwnership(snippetId, userId)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !bool {
		app.clientError(w, http.StatusForbidden)
		return
	}

	snippet, err := app.snippets.Get(snippetId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNoRecord):
			http.NotFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Form = SnippetEditForm{
		ID:      snippet.ID.Hex(),
		Title:   snippet.Title,
		Content: snippet.Content,
		Expires: int(snippet.Expires.Sub(snippet.Created).Hours() / 24),
	}

	app.render(w, r, http.StatusOK, "edit.tmpl.html", data)
}

func (app *application) snippetEditPost(w http.ResponseWriter, r *http.Request) {
	var form SnippetEditForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "edit.tmpl.html", data)
		return
	}

	err = app.snippets.Update(form.ID, form.Title, form.Content, form.Expires)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNoRecord):
			http.NotFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully updated!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%s", form.ID), http.StatusSeeOther)
}
