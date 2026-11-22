package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"golang.org/x/term"
)

const version = "1.2.3"

func main() {
	// --- Флаги ---
	var (
		showVersion bool
		showHelp    bool
		dryRun      bool
	)
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&showVersion, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "print this help and exit")
	flag.BoolVar(&showHelp, "h", false, "print this help and exit (shorthand)")
	flag.BoolVar(&dryRun, "dry-run", false, "print commands without executing (task --dry)")

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

	// --- SIGINT/SIGTERM ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		restoreTerminal()
		clearScreen()
		fmt.Println("Interrupted.")
		os.Exit(130)
	}()

	reader := bufio.NewReader(os.Stdin)

	tasks, err := loadTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get tasks: error %v\n", err)
		waitEnterOrEsc(reader)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Println("There are no tasks in this Taskfile")
		waitEnterOrEsc(reader)
		os.Exit(0)
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Println("Not a terminal")
		os.Exit(130)
	}
	runInteractive(tasks, dryRun)
}

func printUsage() {
	colorPrintf("%s\n\n", bold("taskmenu — interactive menu for Taskfile.yml"))

	colorPrintf("%s\n", bold("USAGE"))
	colorPrintf("  taskmenu [--dry-run]   run interactive menu\n")
	colorPrintf("  taskmenu -h|--help     print this help\n")
	colorPrintf("  taskmenu -v|--version  print version\n")

	colorPrintf("\n%s\n", bold("DESCRIPTION"))
	colorPrintf("  Reads the task list from `task --list-all --json` and shows an\n")
	colorPrintf("  interactive menu. Move with arrows, select with space and press\n")
	colorPrintf("  Enter to run.\n")

	colorPrintf("\n%s\n", bold("FLAGS"))
	colorPrintf("  --dry-run   print commands without executing them (task --dry)\n")

	colorPrintf("\n%s\n", bold("VERSION"))
	colorPrintf("  %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
}

func printVersion() {
	colorPrintf("taskmenu %s (%s/%s)\n",
		version, runtime.GOOS, runtime.GOARCH)
}
