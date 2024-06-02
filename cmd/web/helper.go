package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/burstman/baseRegistry/cmd/web/internal/data"
	"github.com/go-playground/form/v4"
)

// decodePostForm decodes the form data from the given HTTP request and stores the
// result in the provided destination. If there is an error parsing the form or
// decoding the data, the error is returned.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
		return err
	}
	return nil
}

// newTemplateData creates a new templateData struct with the flash message
// from the session manager.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
	}
}
func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "authenticatedUserID")
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// render renders a page using the provided template and data. It writes the rendered
// content to the provided http.ResponseWriter with the given status code.
// If the template does not exist in the application's template cache, it calls
// serverError to handle the error.
// If there is an error executing the template, it also calls serverError.
func (app *application) render(w http.ResponseWriter, page string, status int, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("template does not exist")
		app.serverError(w, err)
		return
	}
	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errlog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) InsertProject(p data.Project) (uint, error) {
	idProject, err := app.projects.InsertProject(p)
	if err != nil {
		return 0, err
	}
	fmt.Println("Project inserted with ID:", idProject)
	return idProject, nil
}

// InsertTask inserts a new task and returns the task ID.
func (app *application) InsertTask(t data.Task) (uint, error) {
	taskID, err := app.projects.InsertTask(t)
	if err != nil {
		return 0, err
	}
	fmt.Println("Task inserted with ID:", taskID)
	return taskID, nil
}

// AddComment adds a new comment to a task.
func (app *application) AddComment(c data.Comment) error {
	_, err := app.projects.AddComment(c)
	if err != nil {
		return err
	}
	fmt.Println("Comment added")
	return nil
}

// AddAttachment adds a new attachment to a task.
func (app *application) AddAttachment(a data.Attachment) error {
	_, err := app.projects.AddAttach(a)
	if err != nil {
		return err
	}
	fmt.Println("Attachment added")
	return nil
}
func (app *application) GetUserID(username string) (uint, error) {
	id, err := app.projects.GetIDFromUserName(username)
	return uint(id), err
}
