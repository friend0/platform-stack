package server

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/spf13/viper"
	"github.com/xo/dburl"
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
	RDB    *redis.UniversalClient
}

func NewServer() (s *ServerBase) {
	return &ServerBase{
		Engine: SetupEngine(),
	}
}

func (s *ServerBase) InitViper() error {
	viper.AutomaticEnv()
	s.Viper = viper.GetViper()
	return nil
}

func ParsePGUrl(url string) (pgurl string, err error) {
	dbu, err := dburl.Parse(url)
	if err != nil {
		return "", err
	}
	return dburl.GenPostgres(dbu)
}

func (s *ServerBase) InitDatabase(url string) (err error) {
	pgurl, err := ParsePGUrl(url)
	if err != nil {
		return err
	}
	db, err := sqlx.Connect("postgres", pgurl)
	if err != nil {
		return err
	}
	s.DB = db
	return err
}

func (s *ServerBase) InitRedis(addr, pass string, rdb int) (err error) {
	ruc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{addr},
		Password: pass,
		DB:       rdb,
	})
	s.RDB = &ruc
	return nil
}

func (s *ServerBase) InitGQLClient(gqlhost string) (err error) {
	s.GQL = graphql.NewClient(gqlhost)
	return nil
}

func (s *ServerBase) InitHTTPClient() (err error) {
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
	s.Client = &http.Client{
		Transport: tr,
		Timeout:   2 * time.Second,
	}
	return nil
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
