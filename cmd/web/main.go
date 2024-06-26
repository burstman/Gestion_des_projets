package main

import (
	// go built-in packages
	"context"
	"database/sql"
	"encoding/gob"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	// add aliases texternao pacakges (internal / external )

	// internal pacakges
	chatApi "github.com/burstman/baseRegistry/cmd/web/internal/chatApi"
	"github.com/burstman/baseRegistry/cmd/web/internal/data"
	"github.com/go-playground/form/v4"

	// external packages
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	addr      string
	staticDir string
	db        struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

var cfg config

type application struct {
	projects        *data.ProjectManager
	userData        *data.UserDB
	chatData        *data.ChatData
	errlog, infolog *log.Logger
	templateCache   map[string]*template.Template
	sessionManager  *scs.SessionManager
	formDecoder     *form.Decoder
	sendRecive      chatApi.SenderReceiver
}

func init() {
	gob.Register([]*ChatHistory{})
}

func main() {
	errlog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)
	infolog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	templateCache, err := newTemplateCache()
	if err != nil {
		errlog.Fatal(err)
	}

	//flags
	flag.StringVar(&cfg.addr, "addr", os.Getenv("REGISTRY_ADDR"), "HTTP Network Addess")
	flag.StringVar(&cfg.staticDir, "static-dir", os.Getenv("REGISTRY_STATIC_DIR"), "Path to static asset")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DSN_PROJECT_Management"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()
	db, err := openDB(cfg)
	if err != nil {
		errlog.Fatal(err)
	}
	//Session Manager
	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true //important

	formDecoder := form.NewDecoder()

	chat := chatApi.NewSenderReceive("http://localhost:8000/send_data")

	app := &application{
		projects:       &data.ProjectManager{DB: db},
		userData:       &data.UserDB{DB: db},
		chatData:       &data.ChatData{DB: db},
		errlog:         errlog,
		infolog:        infolog,
		templateCache:  templateCache,
		sessionManager: sessionManager,
		formDecoder:    formDecoder,
		sendRecive:     chat,
	}

	defer db.Close()
	infolog.Printf("dabase connection pool established at %s\n", cfg.db.dsn)
	srv := &http.Server{
		Addr:         cfg.addr,
		ErrorLog:     app.errlog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	//log info server
	infolog.Printf("server starting at port %s\n", cfg.addr)
	err = srv.ListenAndServe()
	errlog.Fatal(err)
}

// put it into seperate package (maybe?)
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
