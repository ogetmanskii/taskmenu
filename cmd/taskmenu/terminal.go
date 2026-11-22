package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

const (
	// reservedRows — число строк под подсказку, notice и справку, которые
	// резервируются под список, чтобы окно не выталкивало их за экран.
	reservedRows = 5
	// fallbackHeight используется, если размер терминала определить не удалось:
	// лучше показать окно фиксированной высоты, чем вывести весь список за
	// пределы экрана.
	fallbackHeight = 20
)

type key int

const (
	keyNone key = iota
	keyUp
	keyDown
	keyLeft
	keyRight
	keyEnter
	keyEsc
	keyCtrlC
	keyCtrlR
	keyRune
	keyEOF
)

// rawFd/rawState хранят состояние терминала в raw-режиме, чтобы обработчик
// сигналов мог восстановить его перед выходом.
var (
	rawFd    int
	rawState *term.State
)

func makeRawTerminal() {
	rawFd = int(os.Stdin.Fd())
	var err error
	rawState, err = term.MakeRaw(rawFd)
	if err != nil {
		fmt.Printf("could not go into raw mode: %v", err)
		os.Exit(130)
	}
}

func restoreTerminal() {
	if rawState != nil {
		_ = term.Restore(rawFd, rawState)
		rawState = nil
	}
	showCursor()
}

// startResizeWatcher запускает горутину, которая перерисовывает меню при
// изменении размера терминала. Возвращённая функция останавливает наблюдение
// и дожидается завершения горутины, чтобы она не писала на экран во время
// обработки нажатия.
func startResizeWatcher(st *menuState) func() {
	stop := make(chan struct{})
	done := make(chan struct{})
	lastW, lastH, _ := terminalSize()

	go func() {
		defer close(done)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				w, h, ok := terminalSize()
				if !ok || (w == lastW && h == lastH) {
					continue
				}
				lastW, lastH = w, h
				st.render()
			}
		}
	}()

	return func() {
		close(stop)
		<-done
	}
}

// terminalSize возвращает размер видимой области терминала. На Windows
// GetConsoleScreenBufferInfo требует handle экранного буфера, поэтому сначала
// пробуем stdout, а stdin используем как запасной вариант.
func terminalSize() (width, height int, ok bool) {
	for _, fd := range []int{int(os.Stdout.Fd()), rawFd} {
		if w, h, err := term.GetSize(fd); err == nil && h > 0 {
			return w, h, true
		}
	}
	return 0, 0, false
}

// visibleRows возвращает количество строк списка, помещающихся в терминал.
// Если размер терминала неизвестен, используем консервативную высоту, чтобы
// длинный список всё равно прокручивался, а курсор оставался на экране.
func visibleRows(total int) int {
	height := fallbackHeight
	if _, h, ok := terminalSize(); ok {
		height = h
	}
	rows := height - reservedRows
	if rows < 1 {
		rows = 1
	}
	if rows > total {
		rows = total
	}
	return rows
}

// readKey читает одно нажатие. Стрелки приходят как escape-последовательности
// ESC [ A/B.
func readKey(r *bufio.Reader) (key, rune) {
	b, err := r.ReadByte()
	if err != nil {
		return keyEOF, 0
	}

	switch b {
	case '\r', '\n':
		return keyEnter, 0
	case 0x03: // Ctrl+C
		return keyCtrlC, 0
	case 0x12: // Ctrl+R
		return keyCtrlR, 0
	case 0x1b: // ESC
		// Стрелки приходят целиком, поэтому одиночный ESC не должен ждать.
		if r.Buffered() >= 2 {
			b2, _ := r.ReadByte()
			b3, _ := r.ReadByte()
			if b2 == '[' {
				switch b3 {
				case 'A':
					return keyUp, 0
				case 'B':
					return keyDown, 0
				case 'C':
					return keyRight, 0
				case 'D':
					return keyLeft, 0
				}
			}
		}
		return keyEsc, 0
	default:
		if b >= 0x20 && b <= 0x7e {
			return keyRune, rune(b)
		}
	}
	return keyNone, 0
}

func waitEnterOrEsc(reader *bufio.Reader) {
	defer restoreTerminal()
	makeRawTerminal()
	hideCursor()

	colorPrintf("%s", gray("Press Enter or Esc to continue"))
	for {
		key, _ := readKey(reader)
		if key == keyEnter || key == keyEsc {
			return
		}
	}
}
