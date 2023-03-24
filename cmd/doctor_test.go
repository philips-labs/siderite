package cmd

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DoctorCommand(t *testing.T) {
	cmd := NewDoctorCmd()
	b := bytes.NewBufferString("")
	cmd.SetArgs([]string{"-h"})
	cmd.SetOut(b)
	_ = cmd.Execute()

	out, err := io.ReadAll(b)
	if !assert.Nil(t, err) {
		return
	}
	if assert.Less(t, 0, len(out)) {
		return
	}
}
