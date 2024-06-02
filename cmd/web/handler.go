package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/burstman/baseRegistry/cmd/web/internal/data"
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

func (app *application) SendchatMessage(w http.ResponseWriter, r *http.Request) {

	var form userChatForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}
	userID, ok := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)
	if !ok {
		app.serverError(w, fmt.Errorf("failed to convert authenticatedUserID to int"))
		return
	}
	chatHistories, ok := app.sessionManager.Get(r.Context(), "chatMessage").([]*ChatHistory)
	if !ok {
		app.errlog.Println("failed to convert chatMessage to []*ChatHistory")
		return
	}
	userData, err := app.userData.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	chatHistories = append(chatHistories, &ChatHistory{ChatUser: userData.Name,
		ChatMessage: form.Message,
		ChatTime:    time.Now().Format("15:04")})

	chatBotResponse, err := app.sendRecive.SendReceive(userID, form.Message)
	if err != nil {
		app.serverError(w, err)
		return
	}
	var chatOrder *data.ChatOrder
	if chatBotResponse.Id != 0 {
		chatOrder, err = app.chatData.RetrieveUserOrder(chatBotResponse.Id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		fmt.Println(chatOrder)
		chatHistories = append(chatHistories, &ChatHistory{ChatUser: "Bot",
			ChatMessage: fmt.Sprintf("%s : %s : %s : %s : %s", chatOrder.Intent, chatOrder.Projects,
				chatOrder.Tasks, chatOrder.Users, chatBotResponse.Message),
			ChatTime: time.Now().Format("15:04")})
		fmt.Println(len(chatHistories))
		app.sessionManager.Put(r.Context(), "chatMessage", chatHistories)
	}
	var p data.Project
	var t data.Task
	var c data.Comment
	var a data.Attachment

	if chatOrder != nil {
		switch chatOrder.Intent {
		case "create":
			fmt.Println("create")
			for _, project := range chatOrder.Projects {
				fmt.Println("project", project)
				p.Name = &project
				newTypeUserID := uint(userID)
				p.CreatedBy = &newTypeUserID

				// Insert the project
				idProject, err := app.InsertProject(p)
				if err != nil {
					app.serverError(w, err)
					return
				}

				// If there are tasks, insert them
				for _, taskName := range chatOrder.Tasks {
					t.ProjectID = &idProject
					t.Title = &taskName

					taskID, err := app.InsertTask(t)
					if err != nil {
						app.serverError(w, err)
						return
					}

					// If there are comments, insert them
					for _, commentText := range chatOrder.Comments {
						c.TaskID = &taskID
						c.CommentText = &commentText

						if err := app.AddComment(c); err != nil {
							app.serverError(w, err)
							return
						}
					}

					// If there are attachments, insert them
					for _, username := range chatOrder.Users {
						uploadedByID, err := app.GetUserID(username)
						if err != nil {
							app.serverError(w, err)
							return
						}

						a.TaskID = &taskID
						a.UploadedBy = &uploadedByID

						if err := app.AddAttachment(a); err != nil {
							app.serverError(w, err)
							return
						}
					}
				}
			}
		}
	} else {
		fmt.Println("chatorder is nil")
	}

	http.Redirect(w, r, fmt.Sprintf("/tasks/view/%d", userID), http.StatusSeeOther)
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
	projects, err := app.projects.GetAllProjects()
	if err != nil {
		app.serverError(w, err)
		return
	}
	chathistory, ok := app.sessionManager.Get(r.Context(), "chatMessage").([]*ChatHistory)
	if !ok {
		app.serverError(w, fmt.Errorf("enable to extract chat history"))
	}
	users, err := app.userData.GetAllUserNames()
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.ChatHistories = chathistory
	data.Projects = projects
	data.ListUsers = users
	fmt.Println("data:", data.ChatHistories)

	app.render(w, "tasks.tmpl.html", http.StatusOK, data)
}
