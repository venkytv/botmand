package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

type ExecEngine struct {
	cmd  []string
	env  map[string]string
	comm *EngineQueues
}

func (e *ExecEngine) Start(ctx context.Context) {
	defer e.done()

	cmd := exec.Command(e.cmd[0], e.cmd[1:]...)

	cmd.Env = os.Environ()
	for k, v := range e.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		logrus.Error("Failed to open stdin pipe:", cmd, err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logrus.Error("Failed to open stdout pipe:", cmd, err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logrus.Error("Failed to open stderr pipe:", cmd, err)
		return
	}

	if err = cmd.Start(); err != nil {
		logrus.Error("Failed to start command:", cmd, err)
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Pipe input from WriteQ to the command
	go func() {
		defer stdin.Close()
		for {
			select {
			case t := <-e.comm.WriteQ:
				if _, err := io.WriteString(stdin, t+"\n"); err != nil {
					logrus.Error("Failed to post message to command:", t, err)
				}
			case <-ctx.Done():
				logrus.Debug("Closing stdin channel:", cmd)
				return
			}
		}
	}()

	// Pipe output of comand to ReadQ
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			t := scanner.Text()
			e.comm.ReadQ <- t
		}
		logrus.Debug("Stdout closed:", cmd)
	}()

	// Log command stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			t := scanner.Text()
			logrus.Errorf("Error: %#v: %s", cmd, t)
		}
		logrus.Debug("Stderr closed:", cmd)
	}()

	err = cmd.Wait()
	logrus.Debug("Command done:", cmd, err)
}

func (e *ExecEngine) done() {
	logrus.Debug("Closing engine channels")
	close(e.comm.ReadQ)
	close(e.comm.WriteQ)
}

func NewExecEngine(cmd []string, env map[string]string, comm *EngineQueues) *ExecEngine {
	return &ExecEngine{
		cmd:  cmd,
		env:  env,
		comm: comm,
	}
}
