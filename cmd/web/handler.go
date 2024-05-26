package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	//lastList, err := app.registry.Latest()
	// if err != nil {
	// 	app.serverError(w, err)
	// 	return
	// }

	data := &templateData{
		//WorkersRegistry: lastList,
		IsAuthenticated: app.isAuthenticated(r),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
	}

	app.render(w, "login.tmpl.html", http.StatusOK, data)
}

type userSignupForm struct {
	Name     string `form:"username"`
	Email    string `form:"email"`
	Password string `form:"password"`
	//validator.Validator `form:"-"`
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
	UserName string `form:"username"`
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
	//fmt.Println(form)
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	id, err := app.userData.Athentificate(form.UserName, form.Password)

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
	chatHistories := []*ChatHistory{}
	app.sessionManager.Put(r.Context(), "chatMessage", chatHistories)

	// Use the PopString method to retrieve and remove a value from the session
	// data in one step. If no matching key exists this will return the empty
	// string.
	path := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if path != "" {
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Login successfull")
	http.Redirect(w, r, fmt.Sprintf("/tasks/view/%d", id), http.StatusSeeOther)

}

func (app *application) getSignUp(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}

	app.render(w, "sign_up.tmpl.html", http.StatusOK, data)
}

func (app *application) postSignUp(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm
	err := app.decodePostForm(r, &form)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	_, err = app.userData.Register(data.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	})
	if err != nil {
		if errors.Is(err, data.ErrDuplicateName) {
			app.sessionManager.Put(r.Context(), "flash", "Name  all ready exist")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, "login.tmpl.html", http.StatusUnprocessableEntity, data)
			return
		} else if errors.Is(err, data.ErrDuplicateEmail) {
			app.sessionManager.Put(r.Context(), "flash", "Email  all ready exist")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, "login.tmpl.html", http.StatusUnprocessableEntity, data)
			return
		}
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Account created successfully!")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userChatForm struct {
	Message string `form:"message"`
}

func (app *application) chatMessage(w http.ResponseWriter, r *http.Request) {

	var form userChatForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}
	id, ok := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)
	if !ok {
		app.serverError(w, fmt.Errorf("failed to convert authenticatedUserID to int"))
		return
	}
	chatHistories, ok := app.sessionManager.Get(r.Context(), "chatMessage").([]*ChatHistory)
	if !ok {
		app.errlog.Println("failed to convert chatMessage to []*ChatHistory")
		return
	}
	userData, err := app.userData.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	chatHistories = append(chatHistories, &ChatHistory{ChatUser: userData.Name,
		ChatMessage: form.Message,
		ChatTime:    time.Now().Format("15:04")})

	chatBotResponse, err := app.sendRecive.SendReceive(id, form.Message)
	if err != nil {
		app.serverError(w, err)
		return
	}
	chatOrder, err := app.chatData.RetrieveUserOrder(chatBotResponse.Id)

	if err != nil {
		app.serverError(w, err)
		return
	}

	chatHistories = append(chatHistories, &ChatHistory{ChatUser: "Bot",
		ChatMessage: fmt.Sprintf("%s : %s : %s : %s", chatOrder.Intent, chatOrder.Task, chatOrder.Types, chatOrder.User_name),
		ChatTime:    time.Now().Format("15:04")})
	fmt.Println(len(chatHistories))
	app.sessionManager.Put(r.Context(), "chatMessage", chatHistories)
	// if chatOrder.Intent == "create" {

	// }

	http.Redirect(w, r, fmt.Sprintf("/tasks/view/%d", id), http.StatusSeeOther)
}

// userTasksView is an HTTP handler function that retrieves a data registry entry by its ID.
// It extracts the ID from the URL parameters, fetches the corresponding registry entry,
// and renders the view.tmpl.html template with the retrieved data.
// If the ID is invalid or the registry entry is not found, it returns a 404 Not Found response.
// If there is any other error, it logs the error and returns a server error response.
func (app *application) userTasksView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	user, err := app.userData.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	data := app.newTemplateData(r)
	data.User = user
	projects := []*Project{
		{
			Name:        "Project Alpha",
			Description: "First project",
			Status:      "In Progress",
			Deadline:    "2024-06-25",
			Comment:     map[string]string{"Initial setup": "Completed", "Review": "Pending"},
			User:        "John Doe",
			Tasks: []*Task{
				{
					ID:          1,
					Description: "Setup project repository",
					AssignedTo:  "Alice",
					Status:      "Completed",
					DueDate:     "2024-06-25",
					Comments: []Comment{
						{Author: "Alice", Text: "Repository created."},
						{Author: "John", Text: "Reviewed and approved."},
					},
				},
				{
					ID:          2,
					Description: "Initial project setup",
					AssignedTo:  "Bob",
					Status:      "In Progress",
					DueDate:     "2024-06-28",
					Comments: []Comment{
						{Author: "Bob", Text: "Working on the setup."},
					},
				},
			},
		},
		{
			Name:         "Project Beta",
			Description:  "Second project",
			Status:       "Completed",
			Deadline:     time.Date(2024, time.July, 0, 0, 0, 0, 0, time.UTC).Format("25-06-2015"),
			CompleatedAt: time.Date(2024, time.July, 10, 0, 0, 0, 0, time.UTC).Format("25-06-2015"),
			Comment:      map[string]string{"Final presentation": "Completed"},
			User:         "Jane Smith",
			Tasks: []*Task{
				{
					ID:          3,
					Description: "Create project plan",
					AssignedTo:  "Charlie",
					Status:      "Completed",
					DueDate:     "2024-07-05",
					Comments: []Comment{
						{Author: "Charlie", Text: "Project plan created."},
					},
				},
				{
					ID:          4,
					Description: "Conduct team meeting",
					AssignedTo:  "Dave",
					Status:      "Completed",
					DueDate:     "2024-07-08",
					Comments: []Comment{
						{Author: "Dave", Text: "Team meeting held."},
						{Author: "Jane", Text: "Meeting notes shared."},
					},
				},
			},
		},
	}

	data.Projects = projects

	app.render(w, "tasks.tmpl.html", http.StatusOK, data)
}
