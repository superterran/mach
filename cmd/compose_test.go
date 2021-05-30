package cmd

/* https://github.com/KEINOS/Hello-Cobra */

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_composeCmd(t *testing.T) {
	var (
		composeCmd = CreateComposeCmd()
		argsTmp    = []string{}
		buffTmp    = new(bytes.Buffer)

		expect string
		actual string
	)

	composeCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	composeCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := composeCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'restoreCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = ""
	actual = buffTmp.String() // resotre buffer
	assert.Equal(t, expect, actual,
		"Command 'build' with no parameters should produce an empty value.",
	)
}

func Test_composeCmd_Help(t *testing.T) {
	var (
		composeCmd = CreateComposeCmd()
		argsTmp    = []string{"--help"}
		buffTmp    = new(bytes.Buffer)

		expect string
		actual string
	)

	composeCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	composeCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := composeCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'restoreCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = "Usage:"
	actual = buffTmp.String() // resotre buffer
	assert.Contains(t, actual, expect,
		"Command 'help' should show usage",
	)
}

func Test_ComposeMainFlow(t *testing.T) {

	ComposeDirname = "examples/stacks"

	var actual = MainComposeFlow([]string{})

	if actual != nil {
		assert.FailNowf(t, "returned not nil.", "Error msg: %v", actual)
	}
}

func Test_ComposeMainFlowExample(t *testing.T) {

	ComposeDirname = "examples/stacks"

	var actual = MainComposeFlow([]string{"example", "up"})

	if actual != nil {
		assert.FailNowf(t, "returned not nil.", "Error msg: %v", actual)
	}
}
