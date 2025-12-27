package cli

import (
	"fmt"
	"slices"
	"strings"
)

// Command represents CLI commands
type Command interface {
	command()
}

// Login command
type Login struct{}

func (Login) command() {}

// Whoami command
type Whoami struct{}

func (Whoami) command() {}

// Push command
type Push struct{}

func (Push) command() {}

// Pull command
type Pull struct{}

func (Pull) command() {}

// RemoteList command
type RemoteList struct{}

func (RemoteList) command() {}

// RemoteAdd command
type RemoteAdd struct {
	Repo string
}

func (RemoteAdd) command() {}

// RemoteRemove command
type RemoteRemove struct {
	Repo string
}

func (RemoteRemove) command() {}

// Help command
type Help struct{}

func (Help) command() {}

// Unknown command
type Unknown struct {
	Message string
}

func (Unknown) command() {}

// validRemoteSubcommands contains valid remote subcommands
var validRemoteSubcommands = []string{"add", "remove"}

// ParseArgs parses command line arguments and returns a Command
//
// Arguments:
//   - args: Command line arguments (including program name)
//
// Returns:
//   - Command: The parsed command
func ParseArgs(args []string) Command {
	switch len(args) {
	case 0, 1:
		return Help{}
	case 2:
		switch args[1] {
		case "login":
			return Login{}
		case "whoami":
			return Whoami{}
		case "push":
			return Push{}
		case "pull":
			return Pull{}
		case "remote":
			return RemoteList{}
		case "help":
			return Help{}
		default:
			return Unknown{Message: args[1]}
		}
	case 3:
		if args[1] == "remote" {
			subCmd := args[2]
			if slices.Contains(validRemoteSubcommands, subCmd) {
				return Unknown{
					Message: fmt.Sprintf("Missing repository argument. Usage: atat remote %s <owner>/<repo>", subCmd),
				}
			}
			return Unknown{Message: fmt.Sprintf("remote %s", subCmd)}
		}
		return Unknown{Message: args[1]}
	default:
		// 4 or more arguments
		if args[1] == "remote" {
			subCmd := args[2]
			if !slices.Contains(validRemoteSubcommands, subCmd) {
				return Unknown{Message: fmt.Sprintf("remote %s", subCmd)}
			}

			repoArg := args[3]
			parts := strings.Split(repoArg, "/")
			if len(parts) == 2 &&
				parts[0] != "" &&
				parts[1] != "" {
				switch subCmd {
				case "add":
					return RemoteAdd{Repo: repoArg}
				case "remove":
					return RemoteRemove{Repo: repoArg}
				}
			}
			return Unknown{Message: "Invalid repository format. Please use <owner>/<repo>."}
		}
		return Unknown{Message: fmt.Sprintf("%s %s", args[1], args[2])}
	}
}
