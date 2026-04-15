module strict-example

go 1.23.2

require (
	github.com/Mkamono/slogcontext/slogcontext v0.0.0
	github.com/Mkamono/slogcontext/slogcontext/wrapper/strict v0.0.0
)

replace (
	github.com/Mkamono/slogcontext/slogcontext => ../../slogcontext
	github.com/Mkamono/slogcontext/slogcontext/wrapper/strict => ../../slogcontext/wrapper/strict
)
