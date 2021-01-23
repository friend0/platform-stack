package goserver

import (
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/machinebox/graphql"
	"github.com/xo/dburl"
	"net"
	"net/http"
	"time"
)

type Clients struct {
	DB    *sqlx.DB
	GQL   *graphql.Client
	HTTP  *http.Client
	RDB   *redis.UniversalClient
}

func ParsePGUrl(url string) (pgurl string, err error) {
	dbu, err := dburl.Parse(url)
	if err != nil {
		return "", err
	}
	return dburl.GenPostgres(dbu)
}

func InitDatabase(url string) (db *sqlx.DB, err error) {
	pgurl, err := ParsePGUrl(url)
	if err != nil {
		return db, err
	}
	db, err = sqlx.Connect("postgres", pgurl)
	return db, err
}

func InitRedis(addr, pass string, rdb int) (*redis.UniversalClient, error) {
	ruc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{addr},
		Password: pass,
		DB:       rdb,
	})
	return &ruc, nil
}

func InitGQLClient(gqlhost string) (*graphql.Client, error) {
	return graphql.NewClient(gqlhost), nil
}

func InitHTTPClient() (*http.Client, error) {
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

func (s *Clients) Close() {
	if s.DB != nil {
		_ = s.DB.Close()
	}
}