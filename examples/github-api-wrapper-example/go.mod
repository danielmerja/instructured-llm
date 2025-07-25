module github-api-wrapper-example

go 1.24.3

toolchain go1.24.4

replace github.com/tmc/langchaingo => ../..

require github.com/tmc/langchaingo v0.0.0-00010101000000-000000000000

require (
	github.com/google/go-github/v74 v74.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
)
