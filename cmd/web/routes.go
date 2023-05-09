package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"snippetbox.wook.net/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})
	// ui.Files 임베디드 파일 시스템을 가져와서 http.FS 유형으로 변환하여
	// http.FileSystem 인터페이스를 만족하도록 합니다.
	// 그런 다음 이를 http.FileServer() 함수에 전달하여 파일 서버 핸들러를 생성합니다.
	fileServer := http.FileServer(http.FS(ui.Files))

	// 정적 파일은 ui.Files 임베디드 파일 시스템의 "static" 폴더에 포함되어 있습니다.
	//예를 들어, CSS 스타일시트는 "static/css/main.css"에 있습니다.
	// 즉, 이제 더 이상 요청 URL에서 접두사를 제거할 필요가 없으며 /static/으로 시작하는
	// 모든 요청은 파일 서버로 직접 전달하면 해당 정적 파일이 존재하는 한 제공될 수 있습니다.
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)
	// router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// Add a new GET /ping route.
	router.HandlerFunc(http.MethodGet, "/ping", ping)

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}
