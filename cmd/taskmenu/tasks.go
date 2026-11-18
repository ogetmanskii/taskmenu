package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Task struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
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

// runTask запускает `task [--dry] <names...> <extraArgs...>` и возвращает код
// возврата процесса. Если процесс не удалось даже запустить, возвращает
// exitCode=-1 и ошибку.
func runTask(names, extraArgs []string, dry bool) (int, error) {
	args := make([]string, 0, len(names)+len(extraArgs)+1)
	if dry {
		args = append(args, "--dry")
	}
	args = append(args, names...)
	args = append(args, extraArgs...)

	cmd := exec.Command("task", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}

	// Если это *exec.ExitError — достаём код возврата.
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), nil
	}

	fmt.Printf("\n[could not start task: %v]\n", err)
	return -1, err
}
