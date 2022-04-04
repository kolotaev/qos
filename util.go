package qos

import (
	"fmt"
	"io"
)

func textRespond(conn io.Writer, txt string) {
	conn.Write([]byte(txt + "\n"))
}

func okRespond(conn io.Writer) {
	conn.Write([]byte("OK\n"))
}

func errorRespond(conn io.Writer, err error) {
	conn.Write([]byte(fmt.Sprintf("Error: %s\n", err)))
}
