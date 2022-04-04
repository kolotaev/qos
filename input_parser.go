package qos

import (
	"errors"
	"fmt"
	"strings"
)

// Commands language
var commandLanguageRules = map[string]struct {
	ActionMark  string
	ArgsCount   int
	IsHalt      bool
	Description string
}{
	"STOP":     {"STOP", 0, true, "Stop server"},
	"FILE":     {"FILE", 1, false, "Download a file (args: file_name)"},
	"THROTTLE": {"THROTTLE", 2, false, "Enable or disable throttling for a server (args: yes/no)"},
	"SLIMIT":   {"SLIMIT", 2, false, "Set bandwidth limit per server (args: srv_name limit_number)"},
	"CLIMIT":   {"CLIMIT", 2, false, "Set bandwidth limit per connection (args: srv_name limit_number)"},
}

// Command convenient command object from a parsed text command
type Command struct {
	Action string
	Args   []string
	IsHalt bool
}

// GetArg return command argument by number
func (c *Command) GetArg(number int) string {
	if len(c.Args) == 0 {
		return ""
	}
	if number >= len(c.Args) {
		return ""
	}
	return c.Args[number]
}

// ParseInput parse raw string input into a command object
func ParseInput(input string) (*Command, error) {
	input = strings.TrimSpace(input)
	inputs := strings.Split(input, " ")

	if len(inputs) == 0 {
		return nil, errors.New("command can not be an empty string")
	}
	if _, ok := commandLanguageRules[inputs[0]]; !ok {
		return nil, fmt.Errorf("received unknown command: `%s`", input)
	}

	commandRule := commandLanguageRules[inputs[0]]
	if len(inputs)-1 != commandRule.ArgsCount {
		return nil, fmt.Errorf(
			"command arguments count mismatch. Got: %d. Want: %d", len(inputs)-1, commandRule.ArgsCount,
		)
	}

	cmd := &Command{
		Action: commandRule.ActionMark,
		Args:   []string{},
		IsHalt: commandRule.IsHalt,
	}

	for i := 1; i <= commandRule.ArgsCount; i++ {
		cmd.Args = append(cmd.Args, strings.TrimSpace(inputs[i]))
	}

	return cmd, nil
}
