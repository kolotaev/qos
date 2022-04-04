package qos

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextRespond(t *testing.T) {
	w := bytes.NewBuffer([]byte(""))
	textRespond(w, "Foo bar")
	assert.Equal(t, "Foo bar\n", w.String())
}

func TestOkRespond(t *testing.T) {
	w := bytes.NewBuffer([]byte(""))
	okRespond(w)
	assert.Equal(t, "OK\n", w.String())
}

func TestErrorRespond(t *testing.T) {
	w := bytes.NewBuffer([]byte(""))
	errorRespond(w, fmt.Errorf("Err occurred %s", "here"))
	assert.Equal(t, "Error: Err occurred here\n", w.String())
}
