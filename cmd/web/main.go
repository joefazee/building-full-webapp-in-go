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

	"abahjoseph.com/snippetbox/pkg/models/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

type contextKey string

const (
	contextKeyIsAuthenticated = contextKey("isAuthenticated")
	authenticatedUserID       = "authenticatedUserID"
)

type application struct {
	errLog        *log.Logger
	infoLog       *log.Logger
	snippets      SnippetsInterface
	templateCache map[string]*template.Template
	session       *sessions.Session
	users         UsersInterface
}

func main() {

	addr := flag.String("addr", ":4000", "HTTP network address")
	dns := flag.String("dns", "root:root@tcp(localhost:8889)/snippetbox?parseTime=true", "MySQL data source name")
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// setup mysql
	db, err := openDB(*dns)
	if err != nil {
		errLog.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(100) // Better left to sql.DB to manage
	db.SetMaxIdleConns(5)

	// Template cache
	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errLog.Fatal(err)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true
	session.SameSite = http.SameSiteStrictMode

	tlsConfig := &tls.Config{
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		PreferServerCipherSuites: true,
	}

	app := &application{
		infoLog:       infoLog,
		errLog:        errLog,
		snippets:      &mysql.SnippetModel{DB: db},
		users:         &mysql.UserModel{DB: db},
		templateCache: templateCache,
		session:       session,
	}

	// HTTP Server Config
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     app.infoLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	app.infoLog.Printf("Starting server on port %s\n", *addr)
	if err := srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem"); err != nil {
		app.errLog.Fatal(err)
	}
}

func openDB(dns string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dns)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
