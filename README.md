# gomoddirectivecomments: go.mod directive comments

[![Go Reference](https://pkg.go.dev/badge/github.com/AkihiroSuda/gomoddirectivecomments.svg)](https://pkg.go.dev/github.com/AkihiroSuda/gomoddirectivecomments)

Package `gomoddirectivecomments` provides a parser for Go module directive comments
that specify "policies" for module dependencies.

This package will be used by:
- [gomodjail: jail for Go modules](https://github.com/AkihiroSuda/gomodjail)
- [gosocialcheck: social reputation checker for Go modules](https://github.com/AkihiroSuda/gosocialcheck)


## Example

```go-module
module example.com/main

go 1.23

require example.com/dependency v1.2.3 // gomodjail:confined
```

```go
mod, _ := modfile.Parse("go.mod", []byte(goMod), nil)
policies, _ := gomoddirectivecomments.Parse(mod, "gomodjail", "unconfined")
// policies = {"example.com/dependency": "confined"}
```

See [`example_test.go`](./example_test.go) for the full code.


## Further examples

```go-module
// gomodjail:confined
module example.com/foo

go 1.23

require (
        example.com/mod100 v1.2.3
        example.com/mod101 v1.2.3 // gomodjail:unconfined
        example.com/mod102 v1.2.3
        // gomodjail:unconfined
        example.com/mod103 v1.2.3
)

require (
        // gomodjail:unconfined
        example.com/mod200 v1.2.3 // indirect
        example.com/mod201 v1.2.3 // indirect
)

//gomodjail:unconfined
require (
        example.com/mod300 v1.2.3
        example.com/mod301 v1.2.3 // gomodjail:confined
        example.com/mod302 v1.2.3
)
```

This makes the following modules confined: `mod100`, `mod102`, `mod201`, and `mod301`.

The version numbers are ignored.