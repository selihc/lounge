package lounge

import (
	"os"
	"testing"
)

func TestLog(t *testing.T) {

	l := NewDefaultLog(WithOutput(os.Stderr), WithDebugEnabled())

	l.Infof("testing log: %s", "thigns")

	l = l.With(map[string]string{"cool": "stuff"})

	l.Infof("testing log 2: %s", "thigns")
}
