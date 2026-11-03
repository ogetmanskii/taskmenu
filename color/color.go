package color

import (
	"fmt"

	"github.com/mattn/go-colorable"
)

const (
	ColorReset = "\x1b[0m"
	ColorCyan  = "\x1b[36m"
	ColorGray  = "\x1b[90m"
	ColorGreen = "\x1b[32m"
	ColorRed   = "\x1b[31m"
	ColorBold  = "\x1b[1m"
)

var out = colorable.NewColorableStdout()

func Red(str string) string {
	return ColorRed + str + ColorReset
}

func Cyan(str string) string {
	return ColorCyan + str + ColorReset
}

func Gray(str string) string {
	return ColorGray + str + ColorReset
}

func Green(str string) string {
	return ColorGreen + str + ColorReset
}

func Println(str string) {
	fmt.Fprintln(out, str)
}

func Printf(str string, a ...any) {
	fmt.Fprintf(out, str, a...)
}
