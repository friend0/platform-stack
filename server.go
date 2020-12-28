package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

type ServerBase struct {
	Engine *gin.Engine
	DB     *sqlx.DB
	GQL    *graphql.Client
	Client *http.Client
	Viper  *viper.Viper
	RDB    redis.UniversalClient
}

func NewServer() (s *ServerBase) {
	return &ServerBase{
		Engine: SetupEngine(),
	}
}

func (s *ServerBase) InitDependencies(dependencies ...string) {
	for _, dep := range dependencies {
		switch dep {
		case "database":
			db, err := SetupDatabase()
			if err != nil {
				break
			}
			s.DB = db
		case "gql":
			gql, err := SetupGQLClient()
			if err != nil {
				break
			}
			s.GQL = gql
		case "client":
			client, err := SetupHTTPClient()
			if err != nil {
				break
			}
			s.Client = client
		case "viper":
			viper.AutomaticEnv()
			s.Viper = viper.GetViper()
		}
	}
}

func (s *ServerBase) InitFunctions(dependencies ...func() error) (err error) {
	for _, dep := range dependencies {
		err = dep()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ServerBase) InitDB() error {
	db, err := SetupDatabase()
	if err != nil {
		return err
	}
	s.DB = db
	return nil
}

func (s *ServerBase) InitRDB() error {
	rdb, err := SetupRedis()
	if err != nil {
		return err
	}
	s.RDB = rdb
	return nil
}

func (s *ServerBase) InitGQLClient() error {
	gql, err := SetupGQLClient()
	if err != nil {
		return err
	}
	s.GQL = gql
	return nil
}

func (s *ServerBase) InitHTTPClient() error {
	client, err := SetupHTTPClient()
	if err != nil {
		return err
	}
	s.Client = client
	return nil
}

func (s *ServerBase) InitViper() error {
	viper.AutomaticEnv()
	s.Viper = viper.GetViper()
	return nil
}

func (s *ServerBase) InjectDependencies(db *sqlx.DB) {
	if db != nil {
		s.DB = db
	}
}

func SetupDatabase() (db *sqlx.DB, err error) {
	// todo: connection string
	host := viper.GetString("API_DB_HOST")
	port := viper.GetString("API_DB_PORT")
	user := viper.GetString("API_DB_USERNAME")
	password := viper.GetString("API_DB_PASSWORD")
	name := viper.GetString("API_DB_NAME")
	mode := viper.GetString("API_DB_MODE")

	// todo: template string
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, name, mode)

	return sqlx.Connect("postgres", connectionString)
}

func SetupRedis() (db redis.UniversalClient, err error) {
	viper.SetDefault("REDIS_DB", 0) // default DB
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{viper.GetString("REDIS_ADDR")},
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	})
	return rdb, nil
}

func SetupGQLClient() (db *graphql.Client, err error) {
	gqlServer := viper.GetString("API_GQL_HOST")

	return graphql.NewClient(gqlServer), nil
}

func SetupHTTPClient() (client *http.Client, err error) {

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 5 * time.Second,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   2 * time.Second,
	}, nil

}

func SetupEngine() (router *gin.Engine) {
	router = gin.Default()
	return router

}

func (s *ServerBase) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Engine.ServeHTTP(rr, req)
	return rr
}

func (s *ServerBase) Close() {
	if s.DB != nil {
		_ = s.DB.Close()
	}
}

func (s *ServerBase) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Engine.ServeHTTP(w, r)
}
