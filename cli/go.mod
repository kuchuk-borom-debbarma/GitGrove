module github.com/kuchuk-borom-debbarma/GitGrove/cli

go 1.25.4

replace github.com/kuchuk-borom-debbarma/GitGrove/core => ../core

require github.com/kuchuk-borom-debbarma/GitGrove/core v0.0.0-00010101000000-000000000000

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
)
