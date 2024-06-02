package main

import (
	"net/http"

	"github.com/burstman/baseRegistry/cmd/web/ui"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))

	router.Handler(http.MethodGet, "/static/*filepath", fileServer)
	//handler for session Manager
	dynamic := alice.New(app.sessionManager.LoadAndSave, app.authenticated)
	//Handlers
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.getSignUp))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.postSignUp))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.getLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.postLogin))
	router.Handler(http.MethodPost, "/user/sendmessage", dynamic.ThenFunc(app.SendchatMessage))
	router.Handler(http.MethodGet, "/tasks/view/:id", dynamic.ThenFunc(app.userTasksView))
	//protected := dynamic.Append(app.requierAuthentification)

	//router.Handler(http.MethodPost, "/user/message", protected.ThenFunc(app.AddNewChatMessage))
	//router.Handler(http.MethodPost, "/registry/create", protected.ThenFunc(app.addNewDataRegistry))
	//router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.logoutPost))

	standard := alice.New(app.recoverPanic, app.applogRequest, secureHeaders)

	return standard.Then(router)
}
