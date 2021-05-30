package cmd

/* https://github.com/KEINOS/Hello-Cobra */

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_restoreCmd(t *testing.T) {
	var (
		restoreCmd = CreateRestoreCmd()
		argsTmp    = []string{}
		buffTmp    = new(bytes.Buffer)

		expect string
		actual string
	)

	restoreCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	restoreCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := restoreCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'restoreCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = ""
	actual = buffTmp.String() // resotre buffer
	assert.Equal(t, expect, actual,
		"Command 'build' with no parameters should produce an empty value.",
	)
}

func Test_restoreCmd_Help(t *testing.T) {
	var (
		restoreCmd = CreateRestoreCmd()
		argsTmp    = []string{"--help"}
		buffTmp    = new(bytes.Buffer)

		expect string
		actual string
	)

	restoreCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	restoreCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := restoreCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'restoreCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = "Usage:"
	actual = buffTmp.String() // resotre buffer
	assert.Contains(t, actual, expect,
		"Command 'help' should show usage",
	)
}

func Test_restoreCmd_Dryrun(t *testing.T) {

	TestMode = true

	createTempDirectory()

	extractTarball("example-machine")

	var actual = populateMachineDir("example-machine")

	defer os.RemoveAll(tmpDir)

	if actual == false {
		assert.FailNow(t, "mach tag returned as expected, %s", actual)
	}
}
