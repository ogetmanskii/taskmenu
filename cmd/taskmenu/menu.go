package main

import (
	"fmt"
	"strconv"
	"strings"
)

// menuState хранит состояние интерактивного меню.
type menuState struct {
	tasks         []Task
	cursor        int
	offset        int // первый видимый элемент списка (скользящее окно)
	selected      map[int]bool
	selectedOrder []int // индексы задач в порядке выбора
	notice        string
}

// menuEOL завершает строки меню. В raw-режиме на Unix вывод не переводит
// LF в CRLF, поэтому без \r каждая следующая строка начинается с той же
// колонки, где закончилась предыдущая. На Windows CRLF тоже корректен.
const menuEOL = "\r\n"

func (st *menuState) render() {
	// Перерисовка кадра «на месте» (home + затирание строк и хвоста экрана)
	// вместо ESC[2J: полная очистка экрана в VT-терминалах сбрасывает старый
	// кадр в scrollback, из-за чего на каждое нажатие клавиши в истории
	// оставался дубликатель всего меню.
	frameHome()

	total := len(st.tasks)
	if st.cursor < 0 {
		st.cursor = 0
	}
	if st.cursor > total-1 {
		st.cursor = total - 1
	}

	rows := visibleRows(total)
	st.scrollWindow(total, rows)

	end := st.offset + rows
	if end > total {
		end = total
	}

	numbers := st.selectionNumbers()
	numWidth := len(strconv.Itoa(len(numbers)))
	for pos := st.offset; pos < end; pos++ {
		idx := pos
		t := st.tasks[idx]

		cursor := "  "
		if pos == st.cursor {
			cursor = green("▶ ")
		}

		box := strings.Repeat(" ", numWidth)
		selected := false
		if n, ok := numbers[idx]; ok {
			box = cyan(fmt.Sprintf("%*d", numWidth, n))
			selected = true
		}

		name := t.Name
		if selected {
			name = cyan(name)
		}

		colorPrintf("%s%s %s %s%s%s", cursor, box, name, gray(t.Desc), eraseLine(), menuEOL)
	}

	if total == 0 {
		colorPrintf("  no tasks%s%s", eraseLine(), menuEOL)
	}

	if st.notice != "" {
		colorPrintf("%s%s%s", gray(st.notice), eraseLine(), menuEOL)
	}
	colorPrintf("%s%s%s", gray("↑/↓ move · space select · enter run · r reload · q quit"), eraseLine(), menuEOL)
	frameEnd()
}

// buildSelection определяет, какие задачи запускать: отмеченные задачи или
// задачу под курсором.
func (st *menuState) buildSelection() ([]int, bool) {
	if sel := st.selectedIndices(); len(sel) > 0 {
		return sel, true
	}

	if st.cursor >= 0 && st.cursor < len(st.tasks) {
		return []int{st.cursor}, true
	}
	return nil, false
}

// selectedIndices возвращает отмеченные задачи в порядке выбора.
func (st *menuState) selectedIndices() []int {
	if len(st.selectedOrder) == 0 {
		return nil
	}
	return append([]int(nil), st.selectedOrder...)
}

// selectionNumbers возвращает номер запуска (с 1) для каждой выбранной задачи.
func (st *menuState) selectionNumbers() map[int]int {
	numbers := make(map[int]int, len(st.selectedOrder))
	for i, idx := range st.selectedOrder {
		numbers[idx] = i + 1
	}
	return numbers
}

// toggleSelected отмечает или снимает задачу, сохраняя порядок выбора.
func (st *menuState) toggleSelected(idx int) {
	if st.selected[idx] {
		delete(st.selected, idx)
		for i, v := range st.selectedOrder {
			if v == idx {
				st.selectedOrder = append(st.selectedOrder[:i], st.selectedOrder[i+1:]...)
				break
			}
		}
		return
	}
	st.selected[idx] = true
	st.selectedOrder = append(st.selectedOrder, idx)
}

func (st *menuState) moveCursor(d, total int) {
	if total == 0 {
		return
	}
	st.cursor += d
	if st.cursor < 0 {
		st.cursor = total - 1
	}
	if st.cursor > total-1 {
		st.cursor = 0
	}
}

// scrollWindow сдвигает окно так, чтобы курсор всегда был видимым.
func (st *menuState) scrollWindow(total, rows int) {
	if rows <= 0 {
		st.offset = 0
		return
	}
	if st.cursor < st.offset {
		st.offset = st.cursor
	}
	if st.cursor >= st.offset+rows {
		st.offset = st.cursor - rows + 1
	}
	if st.offset < 0 {
		st.offset = 0
	}
	if maxOffset := total - rows; st.offset > maxOffset {
		st.offset = maxOffset
	}
	if st.offset < 0 {
		st.offset = 0
	}
}

func (st *menuState) reload() {
	tasks, err := loadTasks()
	if err != nil {
		st.notice = fmt.Sprintf("could not reload tasks: %v", err)
		return
	}
	st.tasks = tasks
	st.selected = map[int]bool{}
	st.selectedOrder = nil
	st.cursor = 0
	st.offset = 0
	st.notice = "tasks reloaded"
}

// reset вызывается после запуска задач: снимает выделение, но сохраняет
// позицию курсора и окно прокрутки, чтобы после Enter навигация не
// возвращалась к первому элементу.
func (st *menuState) reset() {
	st.selected = map[int]bool{}
	st.selectedOrder = nil
	st.notice = ""
}

// parseLineSelection разбирает строку построчного режима: номера через пробел,
// после `--` — аргументы для задачи.
func parseLineSelection(line string, tasks []Task) ([]string, []string, bool) {
	q, args := splitArgs(line)
	nums, isNum := parseNumbers(q)
	if !isNum {
		colorPrintf("Unknown task: %q\n", line)
		return nil, nil, false
	}

	names := make([]string, 0, len(nums))
	for _, n := range nums {
		if n < 1 || n > len(tasks) {
			colorPrintf("Unknown task: %q\n", line)
			return nil, nil, false
		}
		names = append(names, tasks[n-1].Name)
	}
	return names, args, true
}

func splitArgs(s string) (string, []string) {
	if i := strings.Index(s, "--"); i >= 0 {
		return strings.TrimSpace(s[:i]), strings.Fields(s[i+2:])
	}
	return s, nil
}

func parseNumbers(s string) ([]int, bool) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil, false
	}
	nums := make([]int, 0, len(fields))
	for _, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil {
			return nil, false
		}
		nums = append(nums, n)
	}
	return nums, true
}
