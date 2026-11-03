package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/mattn/go-colorable"
)

const version = "1.0.2"

const (
	colorReset = "\x1b[0m"
	colorCyan  = "\x1b[36m"
	colorGray  = "\x1b[90m"
	colorGreen = "\x1b[32m"
	colorRed   = "\x1b[31m"
	colorBold  = "\x1b[1m"
)

type Task struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
}

var out = colorable.NewColorableStdout()

func red(str string) string {
	return colorRed + str + colorReset
}

func cyan(str string) string {
	return colorCyan + str + colorReset
}

func gray(str string) string {
	return colorGray + str + colorReset
}

func green(str string) string {
	return colorGreen + str + colorReset
}

func println(str string) {
	fmt.Fprintln(out, str)
}

func printf(str string, a ...any) {
	fmt.Fprintf(out, str, a...)
}

func main() {

	// --- Флаги ---
	var (
		showVersion bool
		showHelp    bool
	)
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&showVersion, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "print this help and exit")
	flag.BoolVar(&showHelp, "h", false, "print this help and exit (shorthand)")

	flag.Usage = printUsage
	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}
	if showVersion {
		printVersion()
		os.Exit(0)
	}

	// --- SIGINT ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		clearScreen()
		fmt.Println(out, "Interrupted.")
		os.Exit(130)
	}()

	reader := bufio.NewReader(os.Stdin)

	tasks, err := loadTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get tasks: error %v\n", err)
		pressEnter(reader)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		println("There are no tasks in this Taskfile")
		pressEnter(reader)
		os.Exit(0)
	}

	for {
		clearScreen()

		for i, t := range tasks {
			printf("%s %s %s\n", cyan(strconv.Itoa(i+1)), t.Name, gray(t.Desc))
		}

		printf("%s", green("$ "))

		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		clearScreen()

		parts := strings.Fields(line)
		selected := make([]string, 0, len(parts))
		valid := true

		for _, p := range parts {
			n, err := strconv.Atoi(p)
			if err != nil || n < 1 || n > len(tasks) {
				printf("Unknown task: %q\n", p)
				pressEnter(reader)
				valid = false
				break
			}
			selected = append(selected, tasks[n-1].Name)
		}

		if !valid || len(selected) == 0 {
			continue
		}

		exitCode, runErr := runTask(selected)

		if exitCode != 0 || runErr != nil {
			printf("%s", red("Failed (exit code "+strconv.Itoa(exitCode)+")"))
		}
		pressEnter(reader)
	}
}

func printUsage() {
	printf(`%staskmenu%s — interactive menu for Taskfile.yml

%sUSAGE%s
  taskmenu              run interactive menu
  taskmenu -h|--help    print this help
  taskmenu -v|--version print version

%sDESCRIPTION%s
  Reads the task list from `+"`task --list-all --json`"+` and shows a numbered
  menu. Type a space-separated sequence of numbers to run tasks in order:
      $ 4 2 1
  executes `+"`task task-4 task-2 task-1`"+`. Press Ctrl+C to quit.

%sVERSION%s
  %s (%s/%s)
`,
		colorBold, colorReset,
		colorBold, colorReset,
		colorBold, colorReset,
		colorBold, colorReset,
		version, runtime.GOOS, runtime.GOARCH,
	)
}

func printVersion() {
	printf("taskmenu %s (%s/%s)\n",
		version, runtime.GOOS, runtime.GOARCH)
}

func pressEnter(reader *bufio.Reader) {
	printf("%sPress Enter to continue...%s", colorGray, colorReset)
	_, _ = reader.ReadString('\n')
}

func loadTasks() ([]Task, error) {
	cmd := exec.Command("task", "--list-all", "--json", "--sort", "none")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var list TaskList
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}
	return list.Tasks, nil
}

// runTask запускает `task <args...>` и возвращает код возврата процесса.
// Если процесс не удалось даже запустить, возвращает exitCode=-1 и ошибку.
func runTask(args []string) (int, error) {
	cmd := exec.Command("task", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}

	// Если это *exec.ExitError — достаём код возврата
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), nil
	}

	fmt.Fprintf(out, "\n[could not start task: %v]\n", err)
	return -1, err
}

func clearScreen() {
	printf("\x1b[2J\x1b[H")
}
