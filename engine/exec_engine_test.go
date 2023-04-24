package engine

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	var e Enginer

	qs := NewEngineQueues()

	e = ExecEngine{
		cmd:  "./test.sh",
		env:  map[string]string{},
		comm: &qs,
	}
	go e.Start(context.Background())

	nresp := 200
	qs.WriteQ <- fmt.Sprintf("%d", nresp)

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
