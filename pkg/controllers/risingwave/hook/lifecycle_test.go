package hook

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLifeCycleOption(t *testing.T) {
	var v = 1
	var opt = LifeCycleOption{
		PostReadyFunc: func() error {
			v = 2
			return nil
		},
	}

	opt.PostReadyFunc()
	assert.Equal(t, v, 2)
	if opt.PreUpdateFunc != nil {
		t.Fatal("test failed")
	}
}
