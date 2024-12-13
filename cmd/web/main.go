package main

import (
	"database/sql"
	"encoding/gob"
	"finalProject/data"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = ":8080"

func main() {
	// connect to the database
	db := initDB()
	defer db.Close()
	// create session - connect to redis
	session := initSession()
	// create loggers
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime)
	// create channels
	// createWaitGroup
	wg := sync.WaitGroup{}
	// set up the application config
	app := Config{
		Session:  session,
		DB:       db,
		Wait:     &wg,
		ErrorLog: errorLog,
		InfoLog:  infoLog,
		Models:   data.New(db),
	}
	// listen for signals
	go app.listenForShutdown()

	// set up mail

	// listen for web connections
	app.serve()
}

func (app *Config) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s", webPort),
		Handler: app.routes(),
	}
	app.InfoLog.Println("Starting Web server")
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic("error", err)
	}
}

func initDB() *sql.DB {
	conn := connectToDB()
	if conn == nil {
		log.Panic("Cant connect to DB")
	}
	return conn
}

func connectToDB() *sql.DB {
	counts := 0
	dsn := os.Getenv("DSN")
	for {
		if counts == 3 {
			return nil
		}
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready...")
			counts++
		} else {
			log.Println("DB Connected")
			return connection
		}
		log.Println("Backing off for one second and retrying connection to db")
		time.Sleep(time.Second)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func initSession() *scs.SessionManager {
	gob.Register(data.User{})
	// set up session
	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}

func initRedis() *redis.Pool {
	redisPood := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}
	return redisPood
}

func (app *Config) listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.shutdown()
	os.Exit(0)
}

func (app *Config) shutdown() {
	// perform any cleanup tasks
	app.InfoLog.Println("Would run cleanup tasks....")

	// block until Waitgroup is emnpty
	app.Wait.Wait()

	app.InfoLog.Println("Closing channels and shutting down app")
}
