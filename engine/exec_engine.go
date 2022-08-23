package engine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ExecEngine struct {
	factory string
	cmd     []string
	env     map[string]string
	comm    *EngineQueues
}

func (e ExecEngine) Start(ctx context.Context) {
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

type ExecEngineFactory struct {
	cmd    []string
	config EngineConfig
}

func (eef ExecEngineFactory) Id() string {
	return "EXEC:" + strings.Join(eef.cmd, " ")
}

func (eef ExecEngineFactory) Config() EngineConfig {
	return eef.config
}

func (eef ExecEngineFactory) Create(env map[string]string, comm *EngineQueues) Enginer {
	return ExecEngine{
		factory: eef.Id(),
		cmd:     eef.cmd,
		env:     env,
		comm:    comm,
	}
}

type ExecEngineFactoryLoader struct {
	Dir string
}

func (eel ExecEngineFactoryLoader) Load(ctx context.Context) []EngineFactoryer {
	var engines []EngineFactoryer
	var lock = sync.RWMutex{}
	var wg sync.WaitGroup

	files, err := os.ReadDir(eel.Dir)
	if err != nil {
		logrus.Warnf("Failed to load engines: %s: %s", eel.Dir, err)
		return engines
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for _, f := range files {
		wg.Add(1)

		f := f
		go func() {
			defer wg.Done()

			execPath := path.Join(eel.Dir, f.Name())
			logrus.Debugf("Executing command: %s --config", execPath)
			cmd := exec.CommandContext(ctx, execPath, "--config")

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb

			if err := cmd.Run(); err != nil {
				switch exitError := err.(type) {
				case *exec.ExitError:
					logrus.Warnf("Command exited with error: %s: %d", execPath, exitError.ExitCode())
				case *fs.PathError:
					logrus.Warnf("Error executing file: %s: %s", execPath, exitError.Error())
				default:
					logrus.Warnf("Error running command: %s: %#v", execPath, err)
				}
			} else {
				if errb.Len() > 0 {
					logrus.Warnf("%s --config: error: %s", execPath, errb.String())
				}

				// Defaults
				config := EngineConfig{
					Threaded: true,
				}

				err = yaml.Unmarshal(outb.Bytes(), &config)
				if err != nil {
					logrus.Warnf("Error parsing config: %s: %s: %v", execPath, outb.String(), err)
				}

				eef := ExecEngineFactory{
					cmd:    []string{execPath},
					config: config,
				}

				lock.Lock()
				engines = append(engines, eef)
				lock.Unlock()
			}
		}()
	}

	logrus.Debugf("Waiting for bot engines to load")
	wg.Wait()

	return engines
}

func NewExecEngine(cmd []string, env map[string]string, comm *EngineQueues) *ExecEngine {
	return &ExecEngine{
		cmd:  cmd,
		env:  env,
		comm: comm,
	}
}
