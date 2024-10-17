// internal/cli/cli_test.go

package cli

import (
	"testing"

	"github.com/ngns-io/baxfer/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestNewApp(t *testing.T) {
	log, _ := logger.New("test.log")
	defer log.Close()

	app := NewApp(log)

	assert.Equal(t, "baxfer", app.Name)
	assert.Equal(t, "CLI to help manage storage for database backups", app.Usage)

	// Test that all expected commands are present
	commandNames := []string{"upload", "download", "prune"}
	for _, name := range commandNames {
		command := findCommand(app.Commands, name)
		assert.NotNil(t, command, "Command %s should exist", name)
	}
}

func TestUploadCommand(t *testing.T) {
	log, _ := logger.New("test.log")
	defer log.Close()

	app := NewApp(log)
	uploadCmd := findCommand(app.Commands, "upload")
	assert.NotNil(t, uploadCmd)

	assert.Equal(t, "upload", uploadCmd.Name)
	assert.Contains(t, uploadCmd.Aliases, "u")

	// Test that all expected flags are present
	flagNames := []string{"provider", "region", "bucket", "keyprefix", "backupext", "compress", "non-interactive"}
	for _, name := range flagNames {
		flag := findFlag(uploadCmd.Flags, name)
		assert.NotNil(t, flag, "Flag %s should exist", name)
	}
}

// Helper function to find a command by name
func findCommand(commands []*cli.Command, name string) *cli.Command {
	for _, cmd := range commands {
		if cmd.Name == name {
			return cmd
		}
	}
	return nil
}

// Helper function to find a flag by name
func findFlag(flags []cli.Flag, name string) cli.Flag {
	for _, flag := range flags {
		if flag.Names()[0] == name {
			return flag
		}
	}
	return nil
}
