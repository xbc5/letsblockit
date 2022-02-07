[![logo](.github/logo.svg?raw=true)](https://letsblock.it)

## Remove low-quality content and useless nags, focus on what matters.

[![CI](https://github.com/xvello/letsblockit/actions/workflows/ci.yml/badge.svg)](https://github.com/xvello/letsblockit/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/xvello/letsblockit)](https://goreportcard.com/report/github.com/xvello/letsblockit)

This repository holds the data and server source code for the Let's Block It project:

Check out the project's [about page](https://letsblock.it/help/about)
and [contributing guidelines](https://letsblock.it/help/contributing).


## 

### Minimal example

```go
// rootCommand holds the list of active commands, and embeds owl.Base for common helpers.
type rootCommand struct {
owl.Base
Lower   *lowerCmd `arg:"subcommand:lower" help:"return the text in lowercase"`
}

// lowerCmd holds the arguments for this command.
type lowerCmd struct {
Text string `arg:"positional" help:"text to lowercase"`
}

// Run is the entrypoint for your command, argument values are set in the struct fields.
func (c *lowerCmd) Run(o owl.Owl) {
out := strings.ToLower(c.Text)
require.NotEqual(o, "viper", out, "snakes not allowed here")
o.Println(out)
}

// main just calls owl.RunOwl to parse arguments and start the selected command.
func main() {
owl.RunOwl(new(rootCommand))
}
```

### Going further

- Checkout the [examples folder](https://github.com/xvello/owl/tree/main/examples) to explore the available features
- Read the [go-arg readme](https://github.com/alexflint/go-arg/blob/master/README.md) and
  [reference documentation](https://pkg.go.dev/github.com/alexflint/go-arg) for available argument types
- Read the [owl reference documentation](https://pkg.go.dev/github.com/xvello/owl) for available helpers 