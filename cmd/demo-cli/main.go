package main

import (
	"github.com/Potterli20/golem/internal/version"
	"github.com/Potterli20/golem/pkg/cli"
)

func main() {
	appSettings := cli.AppSettings{
		Name:        "golem test",
		Description: "some fake tool",
		ConfigPath:  "$HOME/.golem/",
		EnvPrefix:   "golem",
		GitVersion:  version.GitVersion,
		GitRevision: version.GitRevision,
		LogLevel:    "debug",
	}

	// Define application level features
	myCli := cli.New[cli.ConfigMock](appSettings)
	defer myCli.Close()

	myCli.Run()
}
