package birc_test

import (
	"fmt"
	"testing"

	"github.com/jpiontek/bitter-irc"
)

func TestCustomLoggerIsDigester(t *testing.T) {
	fn := birc.CustomLogger(nil)
	if _, ok := interface{}(fn).(birc.Digester); !ok {
		t.Error(fmt.Errorf("Logger does not implement Digester"))
	}
}
