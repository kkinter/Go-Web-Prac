package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"snippetbox.wook.net/internal/models"
	"snippetbox.wook.net/ui"
)

type templateData struct {
	CurrentYear     int
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

func humaDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// 템플릿.FuncMap 객체를 초기화하여 전역 변수에 저장합니다.
// 이것은 기본적으로 사용자 정의 템플릿 함수의 이름과 함수 자체 사이의 조회 역할을
// 하는 문자열 키 맵입니다.
var functions = template.FuncMap{
	"humanDate": humaDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// fs.Glob()를 사용하여 'html/pages/*.tmpl' 패턴과 일치하는 ui.Files 임베디드 파일시스템의
	// 모든 파일 경로를 가져옵니다. 이렇게 하면 이전과 마찬가지로 애플리케이션에 대한
	// 모든 '페이지' 템플릿의 슬라이스를 얻을 수 있습니다.
	pages, err := fs.Glob(ui.Files, "html/pages/*.go.tpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// 구문 분석하려는 템플릿의 파일 경로 패턴이 포함된 슬라이스를 만듭니다.
		patterns := []string{
			"html/base.go.tpl",
			"html/partials/*.go.tpl",
			page,
		}

		// ParseFiles() 대신 ParseFS()를 사용하여 ui.Files 임베디드 파일 시스템에서
		// 템플릿 파일을 구문 분석합니다.
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
