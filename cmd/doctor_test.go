package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DoctorCommand(t *testing.T) {
	cmd := NewDoctorCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	err := cmd.Execute()

	if !assert.Nil(t, err) {
		return
	}

}
