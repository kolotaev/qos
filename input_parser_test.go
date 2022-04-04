package qos_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kolotaev/qos"
)

func TestCommand_GetArgExisting(t *testing.T) {
	c := &qos.Command{
		Action: "Foo",
		Args:   []string{"a", "b", "c"},
	}

	assert.Equal(t, "a", c.GetArg(0))
	assert.Equal(t, "b", c.GetArg(1))
	assert.Equal(t, "c", c.GetArg(2))
}

func TestCommand_GetArgMissing(t *testing.T) {
	c := &qos.Command{
		Action: "Foo",
		Args:   []string{},
	}

	assert.Equal(t, "", c.GetArg(0))
	assert.Equal(t, "", c.GetArg(1))
}

func TestParseInput(t *testing.T) {
	cases := []struct {
		input       string
		expected    *qos.Command
		expectError bool
		errorText   string
	}{
		{"", nil, true, "received unknown command: ``"},
		{"  ", nil, true, "received unknown command: ``"},
		{"foobar", nil, true, "received unknown command: `foobar`"},
		{"FILE", nil, true, "command arguments count mismatch. Got: 0. Want: 1"},

		{"STOP", &qos.Command{"STOP", []string{}, true}, false, ""},
		{"  STOP", &qos.Command{"STOP", []string{}, true}, false, ""},
		{"FILE a.txt", &qos.Command{"FILE", []string{"a.txt"}, false}, false, ""},
		{"THROTTLE srv1 33", &qos.Command{"THROTTLE", []string{"srv1", "33"}, false}, false, ""},
		{"SLIMIT srv1 122   ", &qos.Command{"SLIMIT", []string{"srv1", "122"}, false}, false, ""},
		{"CLIMIT 127.0.0.1:88888 500", &qos.Command{"CLIMIT", []string{"127.0.0.1:88888", "500"}, false}, false, ""},
	}

	for i, tc := range cases {
		msg := fmt.Sprintf("Test case #%d", i)
		res, err := qos.ParseInput(tc.input)
		if tc.expectError {
			assert.EqualError(t, err, tc.errorText, msg)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, res, msg)
		}
	}
}
