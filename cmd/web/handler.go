package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/burstman/baseRegistry/cmd/web/internal/data"
	"github.com/burstman/baseRegistry/cmd/web/internal/validator"
	"github.com/julienschmidt/httprouter"
)

// home is an HTTP handler function that retrieves the latest list of workers from the registry
// and renders the "home.tmpl.html" template with the retrieved data.
//
// If there is an error retrieving the latest list of workers, it will call the serverError
// method to handle the error and return a 500 Internal Server Error response.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	println(app.sessionManager.GetInt(r.Context(), "authenticatedUserID"))
	lastList, err := app.registry.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := &templateData{
		WorkersRegistry: lastList,
		IsAuthenticated: app.isAuthenticated(r),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
	}

	app.render(w, "home.tmpl.html", http.StatusOK, data)
}

type userSignupForm struct {
	Name     string `form:"name"`
	Email    string `form:"email"`
	Password string `form:"password"`
	//validator.Validator `form:"-"`
}

// addNewDataRegistryDisplay is an HTTP handler function that renders the "createNew.tmpl.html"
// template with the application's template data. This function is likely used to display a form
// for creating a new data registry entry.
func (app *application) addNewDataRegistryDisplay(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, "createNew.tmpl.html", http.StatusOK, data)
}

// dataRegistryForm represents a form for submitting data to a registry.
// It contains fields for the ID number, name, name of sponsor, place of residence,
// workplace, blood type, and nationality. The Validator field is used for
// validating the form data.
type dataRegistryForm struct {
	IdNumber            string `form:"id_number"`
	Name                string `form:"name"`
	NameOfSponsor       string `form:"Name_of_sponsor"`
	PlaceOfResidence    string `form:"place_of_residence"`
	Workplace           string `form:"workplace"`
	BloodType           string `form:"blood_type"`
	Nationality         string `form:"nationality"`
	validator.Validator `form:"-"`
}

// addNewDataRegistry is an HTTP handler function that processes a form submission to create a new data registry entry.
// It decodes the form data, creates a new registry entry, and redirects the user to the view page for the new entry.
// If there is a duplicate record error, it displays an error message and renders the createNew.tmpl.html template.
// If there is any other error, it logs the error and returns a server error response.
func (app *application) addNewDataRegistry(w http.ResponseWriter, r *http.Request) {
	var form dataRegistryForm
	//data. := dataRegistryForm{

	// 	IdNumber:   worker.IDnumber,
	// 	FieldError: map[string]string{},
	// }
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	id, err := app.registry.Insert(data.RegistryWorker{
		Name:             form.Name,
		IDnumber:         form.IdNumber,
		NameOfSponsor:    form.NameOfSponsor,
		PlaceOfResidence: form.PlaceOfResidence,
		Workplace:        form.Workplace,
		BloodType:        form.BloodType,
		Nationality:      form.Nationality,
	})

	if err != nil {
		if errors.Is(err, data.ErrDuplicateRecord) {
			app.sessionManager.Put(r.Context(), "flash", "Id number all ready exist")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, "createNew.tmpl.html", http.StatusUnprocessableEntity, data)
			return
		}
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Data sent successfully!")

	http.Redirect(w, r, fmt.Sprintf("/registry/view/%d", id), http.StatusSeeOther)

}

// getRegistryId is an HTTP handler function that retrieves a data registry entry by its ID.
// It extracts the ID from the URL parameters, fetches the corresponding registry entry,
// and renders the view.tmpl.html template with the retrieved data.
// If the ID is invalid or the registry entry is not found, it returns a 404 Not Found response.
// If there is any other error, it logs the error and returns a server error response.
func (app *application) getRegistryId(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	rg, err := app.registry.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	flash := app.sessionManager.PopString(r.Context(), "flash")
	data := app.newTemplateData(r)
	data.WorkerRegistry = rg
	data.Flash = flash
	fmt.Println("id not found")
	app.render(w, "view.tmpl.html", http.StatusOK, data)
}

func (app *application) logoutPost(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in.
	if !app.sessionManager.Exists(r.Context(), "authenticatedUserID") {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	// Redirect the user to the application home page.
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

type userLoginForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
	//validator.Validator `form:"-"` //Todo
}

func (app *application) getLogin(w http.ResponseWriter, r *http.Request) {
	println(app.sessionManager.GetInt(r.Context(), "authenticatedUserID"))

	data := app.newTemplateData(r)

	data.Form = userLoginForm{}

	app.render(w, "login.tmpl.html", http.StatusOK, data)
}

func (app *application) postLogin(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in.
	var form userLoginForm
	fmt.Println(form)
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	id, err := app.userData.Athentificate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, data.ErrInvalidCredentials) {
			app.sessionManager.Put(r.Context(), "flash", "Invalid credentials")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, "login.tmpl.html", http.StatusUnprocessableEntity, data)
			return
		}
		app.serverError(w, err)
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	//Add ID to the session Manager
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	// Use the PopString method to retrieve and remove a value from the session
	// data in one step. If no matching key exists this will return the empty
	// string.
	path := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if path != "" {
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Login successfull")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *application) getSignUp(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}

	app.render(w, "signUp.tmpl.html", http.StatusOK, data)
}

func (app *application) postSignUp(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm
	err := app.decodePostForm(r, &form)
	fmt.Println(form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	_, err = app.userData.Register(data.UserAuth{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	})
	if err != nil {
		if errors.Is(err, data.ErrDuplicateName) {
			app.sessionManager.Put(r.Context(), "flash", "Name  all ready exist")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, "registring.tmpl.html", http.StatusUnprocessableEntity, data)
			return
		} else if errors.Is(err, data.ErrDuplicateEmail) {
			app.sessionManager.Put(r.Context(), "flash", "Email  all ready exist")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, "registring.tmpl.html", http.StatusUnprocessableEntity, data)
			return
		}
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Account created successfully!")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// chatMessage handles an HTTP request to send a chat message.
// It takes the message ID, message text, and URL to send the data to,
// and passes them to the SendReceive function to handle the message sending.
// If an error occurs, it is logged and a server error response is returned.
func (app *application) chatMessage(w http.ResponseWriter, r *http.Request) {
	message := "I want two  pizza"
	url := "http://localhost:8000/send_data"
	userID := 1
	chatBotResponse, err := app.sendRecive.SendReceive(userID, message, url)
	fmt.Println(chatBotResponse)

	if err != nil {
		app.serverError(w, err)
		return
	}
	if chatBotResponse.Id != 0 {
		chatOrder, err := app.chatData.RetrieveUserOrder(chatBotResponse.Id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		fmt.Println(chatOrder)
	} else {
		fmt.Println(chatBotResponse.Message)
	}
}
