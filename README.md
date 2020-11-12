Platform Go Server
===================

Platform Go Server implements common server functionality built on Gin.

It provides http, Postgres, and GQL clients out of the box, and is easily configurable with Viper. 

To use this server to build your own Go service:

- Define a main cmd like the template given in cmd/go-server/main.go
- Make a `routes.go` or similar file in your module 

```.env
import (
	s "github.com/altiscope/platform-go-server"
	"github.com/gin-gonic/gin"
)

type Server struct {
	*s.ServerBase
}

func NewServer() (*Server) {
	return &Server{
		s.NewServer(),
	}
}

func (s *Server) routes() {
	s.Engine.POST("/briefing", gin.WrapH(s.MyHandler()))
}
```
- Implement your handlers like this:

```.env
func (s *Server) MyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        ... Handler Code ...
	}
}
```

This module was developed using many of the good practices described in Matt Ryer's now seminal Gopher reading ["How I Write Go Services after \[8\] years"](hyttps://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html).
We've found they make writing, testing, debugging, and maintaining Go services much more enjoyable.

## Install
This module is private