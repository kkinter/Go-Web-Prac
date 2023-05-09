package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

// serverError는 오류 메시지와 스택 추적을 errorLog에 기록합니다,
// 그런 다음 일반 500 내부 서버 오류 응답을 사용자에게 보냅니다.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError 헬퍼는 특정 상태 코드와 해당 설명을 사용자에게 전송합니다.
// 사용자에게 전송합니다. 이 책의 뒷부분에서 이를 사용하여 사용자가 보낸 요청에 문제가 있을 때
// 400 "Bad Request"과 같은 응답을 보내는 데 사용하겠습니다.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// 일관성을 위해 notFound 헬퍼도 구현하겠습니다. 이것은 단순히 클라이언트 에러에
// 404 찾을 수 없음 응답을 사용자에게 보내는 클라이언트 오류에 대한 편의 래퍼입니다.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// 새 버퍼 초기화
	buf := new(bytes.Buffer)

	// 템플릿을 버퍼에 바로 쓰지 않고
	// http.ResponseWriter. 오류가 발생하면 serverError() 헬퍼를 호출한 다음 반환합니다.
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
	// 템플릿이 오류 없이 버퍼에 기록되면 안전합니다.
	// HTTP 상태 코드를 http.ResponseWriter에 기록합니다.
	w.WriteHeader(status)
	// 버퍼의 내용을 http.ResponseWriter에 씁니다.
	// 참고: 이번에도 io.Writer를 받는 함수에 http.ResponseWriter를 전달합니다.
	buf.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
	}
}

// 새로운 decodePostForm() 헬퍼 메서드를 생성합니다. 여기서 두 번째 매개변수인 dst는
// 양식 데이터를 디코딩할 대상입니다.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	// 요청에 대해 createSnippetPost 핸들러에서 했던 것과 같은 방식으로 ParseForm()을 호출합니다.
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// 디코더 인스턴스에서 Decode()를 호출하여 첫 번째 매개변수로 대상 대상을 전달합니다.
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// 유효하지 않은 대상 대상을 사용하려고 하면 Decode() 메서드가 *form.InvalidDecoderError 유형의
		// 오류를 반환합니다. 오류를 반환하는 대신 errors.As()를 사용하여
		// 이를 확인하고 패닉을 발생시킵니다.
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
		// 다른 모든 오류의 경우 정상적으로 반환합니다.
		return err
	}

	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}
