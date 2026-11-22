package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

const (
	colorReset = "\x1b[0m"
	colorCyan  = "\x1b[36m"
	colorGray  = "\x1b[90m"
	colorGreen = "\x1b[32m"
	colorRed   = "\x1b[31m"
	colorWhite = "\x1b[37m"
	colorBold  = "\x1b[1m"
)

// Последовательности перерисовки кадра меню.
//
// ESC[2J (очистка всего экрана) в VT-терминалах — Windows Terminal, xterm и
// др. — не просто стирает видимую область: стёртое содержимое сбрасывается
// в scrollback. При перерисовке меню на каждое нажатие клавиши это копит
// дубликаты экрана в истории. Поэтому кадр перерисовывается «на месте»:
// возврат в начало экрана (ESC[H), стирание хвоста строки (ESC[K]) и
// стирание области ниже кадра (ESC[0J) — они не создают записей в истории.
const (
	enterClearSeq = "\x1b[2J\x1b[H" // только один раз при входе в меню
	homeSeq       = "\x1b[1;1H"
	eraseBelowSeq = "\x1b[0J"
	eraseLineSeq  = "\x1b[K"
	hideCursorSeq = "\x1b[?25l"
	showCursorSeq = "\x1b[?25h"
)

var (
	colorOut = colorable.NewColorableStdout()
)

func isTerminalOut() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

func wrapColor(code, str string) string {
	return code + str + colorReset
}

func red(str string) string  { return wrapColor(colorRed, str) }
func cyan(str string) string { return wrapColor(colorCyan, str) }
func gray(str string) string { return wrapColor(colorGray, str) }
func bold(str string) string { return wrapColor(colorBold, str) }

func colorPrintf(str string, a ...any) {
	fmt.Fprintf(colorOut, str, a...)
}

// enterClear полностью очищает экран (ESC[2J). Используется только один раз
// при входе в интерактивное меню: в VT-терминалах очистка сбрасывает
// текущий экран в scrollback — здесь это как раз нужно, чтобы содержимое
// шелла осталось доступным в истории. Для перерисовок кадра применять нельзя.
func enterClear() {
	if isTerminalOut() {
		fmt.Fprint(colorOut, enterClearSeq)
	}
}

// clearScreen стирает видимую область без скролла: возврат в начало экрана и
// стирание вниз. Не создаёт записей в scrollback, в отличие от ESC[2J.
func clearScreen() {
	frameHome()
	frameEnd()
}

// frameHome ставит курсор в левый верхний угол видимой области.
func frameHome() {
	if isTerminalOut() {
		fmt.Fprint(colorOut, homeSeq)
	}
}

// frameEnd стирает всё от курсора до конца экрана — остатки предыдущего
// (более высокого) кадра.
func frameEnd() {
	if isTerminalOut() {
		fmt.Fprint(colorOut, eraseBelowSeq)
	}
}

// eraseLine возвращает последовательность стирания хвоста текущей строки —
// остатки более длинного текста предыдущего кадра на этой же позиции.
func eraseLine() string {
	if isTerminalOut() {
		return eraseLineSeq
	}
	return ""
}

// hideCursor прячет курсор на время показа меню. После перерисовки кадра
// физический курсор остаётся припаркованным за последним выведенным символом,
// и терминал продолжает мигать им, хотя вводить в меню нечего — навигация
// идёт стрелками и пробелом. Скрытый курсор мигать не может.
func hideCursor() {
	if isTerminalOut() {
		fmt.Fprint(colorOut, hideCursorSeq)
	}
}

// showCursor возвращает курсор терминалу. Обязателен на всех выходах из меню,
// иначе шелл останется с невидимым курсором.
func showCursor() {
	if isTerminalOut() {
		fmt.Fprint(colorOut, showCursorSeq)
	}
}
