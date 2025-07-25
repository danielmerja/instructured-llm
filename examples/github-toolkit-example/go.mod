module github-toolkit-example

go 1.24.3

toolchain go1.24.4

replace github.com/tmc/langchaingo => ../..

require github.com/tmc/langchaingo v0.0.0-00010101000000-000000000000

require (
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/google/go-github/v74 v74.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pkoukk/tiktoken-go v0.1.7 // indirect
	go.starlark.net v0.0.0-20230302034142-4b1e35fe2254 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
)
