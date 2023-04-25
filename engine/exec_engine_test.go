package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func SetupTestEngine(ctx context.Context, t *testing.T, engine Enginer, qs *EngineQueues) {
	stdin, stdout, _, err := engine.Setup(ctx)
	assert.Nil(t, err)

	// Pipe input from WriteQ to engine
	go func() {
		defer stdin.Close()
		for {
			select {
			case text := <-qs.WriteQ:
				_, err = io.WriteString(stdin, text)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Pipe output of engine to ReadQ
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			text := scanner.Text()
			qs.ReadQ <- text
		}
	}()

	// Start the engine
	err = engine.Start(ctx)
	assert.Nil(t, err)

	// Wait for engine to finish
	err = engine.Wait(ctx)
	assert.Nil(t, err)

	// Close queues
	close(qs.WriteQ)
	close(qs.ReadQ)
}

func TestExec(t *testing.T) {
	var e Enginer

	qs := NewEngineQueues()

	e = &ExecEngine{
		cmd: "./test.sh",
		env: map[string]string{},
	}
	ctx := context.Background()

	go SetupTestEngine(ctx, t, e, &qs)

	nresp := 200
	qs.WriteQ <- fmt.Sprintf("%d\n", nresp)

	i := 0
	for {
		i += 1
		resp, more := <-qs.ReadQ
		if more {
			assert.Equal(t, fmt.Sprintf("%d", i), resp)
		} else {
			break
		}
	}
	assert.Equal(t, nresp+1, i)
}
