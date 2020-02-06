Building
========

`go build -o /usr/local/bin/stack -v ../platform-stack/main.go`

Testing
=======

Run tests locally:
`go test -v ./cmd`

Update "golden" expected test result files:
`go test -v ./cmd -test.update-golden`
