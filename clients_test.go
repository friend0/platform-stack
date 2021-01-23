package client

import (
	mis "github.com/matryer/is"
	"strings"
	"testing"
)


func TestParsePGUrl(t *testing.T) {
	is := mis.New(t)
	t.Run("test parse PG url", func(t *testing.T) {
		res, err := ParsePGUrl("postgresql://user:pass@localhost/mydatabase?sslmode=disable")
		if err != nil {
			t.Fail()
		}
		splitURL := strings.Split(res, " ")
		is.New(t)
		is.Equal(splitURL[0], "dbname=mydatabase")
	})
}