package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"snippetbox.wook.net/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// form.Decoder 인스턴스에 대한 포인터를 보유하는 formDecoder 필드를 추가합니다.
type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface // Use our new interface type.
	users          models.UserModelInterface    // Use our new interface type.
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP 네트워크 주소")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	// scs.New() 함수를 사용하여 새 세션 관리자를 초기화합니다.
	// 그런 다음 MySQL 데이터베이스를 세션 저장소로 사용하도록 구성하고
	// 수명을 12시간으로 설정합니다(세션이 처음 생성된 후 12시간이 지나면 자동으로 만료되도록).

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	// 세션 쿠키에 보안 속성이 설정되어 있는지 확인합니다.
	// 이 속성을 설정하면 쿠키가 사용자의 웹 브라우저에서만 전송됩니다.
	// 브라우저에서만 전송되며, 보안되지 않은 HTTP 연결을 통해서는 전송되지 않습니다.
	sessionManager.Cookie.Secure = true

	// 그리고 애플리케이션 종속성에 sessionManager를 추가합니다.
	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
	// 서버에서 사용할 기본값이 아닌 TLS 설정을 저장하기 위해 tls.Config 구조체를 초기화합니다.
	// 이 경우 변경하는 것은 커브 기본 설정 값뿐이므로 어셈블리 구현이 있는 타원형 커브만
	// 사용됩니다.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// 서버의 TLSConfig 필드를 방금 생성한 tlsConfig 변수를 사용하도록 설정합니다.
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("%s에서 서버 시작 중", *addr)
	// ListenAndServeTLS() 메서드를 사용하여 HTTPS 서버를 시작합니다.
	// 두 개의 매개변수로 TLS 인증서와 해당 개인 키의 경로를 전달합니다.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

// openDB() 함수는 sql.Open()을 래핑하고
// 주어진 DSN에 대한 sql.DB 연결 풀을 반환합니다.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
