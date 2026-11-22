package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/term"
)

// runInteractive запускает меню в raw-режиме: навигация стрелками, выбор
// пробелом, запуск по Enter, перезагрузка по r/Ctrl+R.
func runInteractive(tasks []Task, dry bool) {
	fd := int(os.Stdin.Fd())
	makeRawTerminal()
	defer restoreTerminal()

	// Один раз при входе: содержимое шелла уезжает в scrollback (это штатное
	// поведение очистки экрана). Дальше все перерисовки кадра идут без
	// ESC[2J, чтобы не плодить дубликаты меню в истории.
	enterClear()
	hideCursor()

	reader := bufio.NewReader(os.Stdin)
	st := &menuState{tasks: tasks, selected: map[int]bool{}}

	for {
		st.render()

		// Пока ждём ввод, следим за размером терминала и перерисовываем
		// меню при его изменении.
		stopWatcher := startResizeWatcher(st)
		k, r := readKey(reader)
		stopWatcher()

		switch k {
		case keyCtrlC:
			restoreTerminal()
			clearScreen()
			fmt.Println("Interrupted.")
			os.Exit(130)

		case keyEOF:
			restoreTerminal()
			return

		case keyEsc:
			restoreTerminal()
			return

		case keyCtrlR:
			st.reload()

		case keyUp:
			st.moveCursor(-1, len(st.tasks))

		case keyDown:
			st.moveCursor(1, len(st.tasks))

		case keyLeft:
			st.unselectAtCursor()

		case keyRight:
			st.selectAtCursor()

		case keyEnter:
			clearScreen()

			indices, ok := st.buildSelection()
			if !ok {
				st.notice = "nothing to run"
				continue
			}

			restoreTerminal()
			var extraArgs []string

			names := make([]string, 0, len(indices))
			for _, i := range indices {
				names = append(names, st.tasks[i].Name)
			}

			code, runErr := runTask(names, extraArgs, dry)
			if code != 0 || runErr != nil {
				colorPrintf("%s\n", red(fmt.Sprintf("Failed (exit code %d)", code)))
			}
			waitEnterOrEsc(reader)

			if ns, err := term.MakeRaw(fd); err == nil {
				rawState = ns
				hideCursor()
			}
			st.reset()

		case keyRune:
			switch r {
			case ' ':
				st.toggleSelectedAtCursor()
			case 'q', 'Q':
				restoreTerminal()
				return
			case 'r', 'R':
				st.reload()
			}
		}
	}
}
