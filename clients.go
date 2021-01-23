package goserver

import (
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/spf13/viper"
	"github.com/xo/dburl"
	"net"
	"net/http"
	"time"
)

type Clients struct {
	DB     *sqlx.DB
	GQL    *graphql.Client
	Client *http.Client
	Viper  *viper.Viper
	RDB    *redis.UniversalClient
}

func (s *Clients) InitViper() error {
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

func (s *Clients) InitDatabase(url string) (err error) {
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

func (s *Clients) InitRedis(addr, pass string, rdb int) (err error) {
	ruc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{addr},
		Password: pass,
		DB:       rdb,
	})
	s.RDB = &ruc
	return nil
}

func (s *Clients) InitGQLClient(gqlhost string) (err error) {
	s.GQL = graphql.NewClient(gqlhost)
	return nil
}

func (s *Clients) InitHTTPClient() (err error) {
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

func (s *Clients) Close() {
	if s.DB != nil {
		_ = s.DB.Close()
	}
}