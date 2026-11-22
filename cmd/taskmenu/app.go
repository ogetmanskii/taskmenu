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
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Printf("could not go into raw mode: %v", err)
		os.Exit(130)
	}
	rawFd = fd
	rawState = oldState
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

		case keyEnter:
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
			if code == 0 && runErr == nil {
				colorPrintf("%s\n", green("Done"))
			} else {
				colorPrintf("%s\n", red(fmt.Sprintf("Failed (exit code %d)", code)))
			}
			waitEnter(reader)

			if ns, err := term.MakeRaw(fd); err == nil {
				rawState = ns
				hideCursor()
			}
			st.reset()

		case keyRune:
			switch r {
			case ' ':
				if st.cursor >= 0 && st.cursor < len(st.tasks) {
					st.toggleSelected(st.cursor)
				}
			case 'q', 'Q':
				restoreTerminal()
				return
			case 'r', 'R':
				st.reload()
			}
		}
	}
}
