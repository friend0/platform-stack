Testing
=======

Run tests locally:
`go test -v ./cmd`

Update "golden" expected test result files:
`ENV=local go test -v ./cmd -test.update-golden`
