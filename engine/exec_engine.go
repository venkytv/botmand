package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// ExecEngine implements the Enginer interface
type ExecEngine struct {
	cmd     string
	env     map[string]string
	execCmd *exec.Cmd
}

func (e *ExecEngine) Setup(ctx context.Context) (io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {
	e.execCmd = exec.Command(e.cmd)

	e.execCmd.Env = os.Environ()
	for k, v := range e.env {
		e.execCmd.Env = append(e.execCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := e.execCmd.StdinPipe()
	if err != nil {
		logrus.Error("Failed to open stdin pipe:", e.execCmd, err)
		return nil, nil, nil, err
	}

	stdout, err := e.execCmd.StdoutPipe()
	if err != nil {
		logrus.Error("Failed to open stdout pipe:", e.execCmd, err)
		return nil, nil, nil, err
	}

	stderr, err := e.execCmd.StderrPipe()
	if err != nil {
		logrus.Error("Failed to open stderr pipe:", e.execCmd, err)
		return nil, nil, nil, err
	}

	logrus.Debugf("Engine setup complete: %#v", e)
	return stdin, stdout, stderr, nil
}

func (e ExecEngine) Start(ctx context.Context) error {
	if err := e.execCmd.Start(); err != nil {
		logrus.Error("Failed to start command:", e.execCmd, err)
		return err
	}
	return nil
}

func (e ExecEngine) Wait(ctx context.Context) error {
	return e.execCmd.Wait()
}

// ExecEngineFactory implements the EngineFactoryer interface
type ExecEngineFactory struct {
	config *Config
}

func (eef ExecEngineFactory) Config() *Config {
	return eef.config
}

func (eef ExecEngineFactory) Create(env map[string]string) Enginer {
	return &ExecEngine{
		cmd: eef.config.Handler,
		env: env,
	}
}

// ExecEngineFactoryLoader implements the EngineFactoryLoader interface
type ExecEngineFactoryLoader struct{}

func (eel ExecEngineFactoryLoader) Load(ctx context.Context, config *Config) EngineFactoryer {
	return ExecEngineFactory{
		config: config,
	}
}
