package main

import (
	"alphalab-cli/pkg/installation"

	"github.com/alecthomas/kong"
)

// // VersionCmd represents the 'version' command
// type VersionCmd struct{}

// // getAppVersion use API to retrive the version of app
// func getAppVersion() (string, error) {
// 	// for testing
// 	return "v0.0.1", nil
// }

// func (v *VersionCmd) Run() error {
// 	// Print the CLI version
// 	fmt.Println("CLI version:", cliVersion)

// 	// Retrieve and print the application version
// 	appVersion, err := getAppVersion()
// 	if err != nil {
// 		return fmt.Errorf("failed to get app version: %w", err)
// 	}
// 	fmt.Println("Application version:", appVersion)
// 	return nil
// }

// CLI struct defines the commands and flags for CLI
type CLI struct {
	// Version VersionCmd              `cmd:"" short:"v" help:"Print CLI & App version information."`
	Install installation.InstallCmd `cmd:"" help:"Install the app."`
	Update  installation.UpdateCmd  `cmd:"" help:"Update the app."`
}

func main() {
	var cli CLI

	// Parse and run the CLI commands
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
