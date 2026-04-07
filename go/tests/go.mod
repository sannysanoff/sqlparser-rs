module github.com/user/sqlparser/tests

go 1.24.2

require (
	github.com/stretchr/testify v1.9.0
	github.com/user/sqlparser v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/user/sqlparser => ..
