package stack

/* https://github.com/KEINOS/Hello-Cobra */

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_stackCmd(t *testing.T) {
	var (
		stackCmd = CreateStackCmd()
		argsTmp  = []string{}
		buffTmp  = new(bytes.Buffer)

		expect string
		actual string
	)

	stackCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	stackCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := stackCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'restoreCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = ""
	actual = buffTmp.String() // resotre buffer
	assert.Equal(t, expect, actual,
		"Command 'build' with no parameters should produce an empty value.",
	)
}

func Test_stackCmd_Help(t *testing.T) {
	var (
		stackCmd = CreateStackCmd()
		argsTmp  = []string{"--help"}
		buffTmp  = new(bytes.Buffer)

		expect string
		actual string
	)

	stackCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	stackCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := stackCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'restoreCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = "Usage:"
	actual = buffTmp.String() // resotre buffer
	assert.Contains(t, actual, expect,
		"Command 'help' should show usage",
	)
}

func Test_StackMainFlowExample(t *testing.T) {

	StacksDirname = "examples/stacks"

	TestMode = true

	var actual = MainStackFlow([]string{"deploy", "example"})

	if actual != nil {
		assert.FailNowf(t, "returned not nil.", "Error msg: %v", actual)
	}
}

func Test_StackMainFlow(t *testing.T) {

	StacksDirname = "examples/stacks"

	TestMode = true

	var actual = MainStackFlow([]string{"deploy"})

	if actual != nil {
		assert.FailNowf(t, "returned not nil.", "Error msg: %v", actual)
	}
}
