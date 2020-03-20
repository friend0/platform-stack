package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
)

type Server struct {
	Router                 *mux.Router
	DB                     *sqlx.DB
	Log                    *log.Logger
	identityServiceIdCache map[string]string
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf(key + " is not set. Falling back to default: " + fallback)
	return fallback
}

func NewServer() (s *Server) {
	s = &Server{
		Router: SetupRouter(),
	}
	s.routes()
	return
}

func InjectEnv() {
	env, ok := os.LookupEnv("ENV")
	if !ok && env != "CI" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

func (s *Server) InitDependencies(dependencies ...string) {
	for _, dep := range dependencies {
		switch dep {
		case "database":
			db, err := SetupDatabase()
			if err != nil {
				break
			}
			s.DB = db
		}
	}
}

func (s *Server) InjectDependencies(db *sqlx.DB, logger *log.Logger) {

	if db != nil {
		s.DB = db
	}
	if logger != nil {
		s.Log = logger
	}
}

func SetupDatabase() (db *sqlx.DB, err error) {
	host := GetEnv("API_DB_HOST", "localhost")
	port := GetEnv("API_DB_PORT", "5432")
	user := GetEnv("API_DB_USERNAME", "psqladmin")
	password := GetEnv("API_DB_PASSWORD", "postgres")
	name := GetEnv("API_DB_NAME", "platformdb1")
	mode := GetEnv("API_DB_MODE", "disable")

	// todo: template string
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, name, mode)
	fmt.Println("CONNECTION STRING: ", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, "SECRET", name, mode))

	db, err = sqlx.Connect("postgres", connectionString)
	return
}

func SetupRouter() (router *mux.Router) {
	router = mux.NewRouter()
	return router

}

func (s *Server) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func (s *Server) Close() {
	_ = s.DB.Close()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
