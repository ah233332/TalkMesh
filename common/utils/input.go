package utils

import (
	"bufio"
	"os"
	"strings"
)

var stdin = bufio.NewReader(os.Stdin)

func Readline() string {
	s, _ := stdin.ReadString('\n')
	return strings.TrimSpace(s)
}
