package backup

/* https://github.com/KEINOS/Hello-Cobra */

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_backupCmd(t *testing.T) {
	var (
		backupCmd = CreateBackupCmd()
		argsTmp   = []string{}
		buffTmp   = new(bytes.Buffer)

		expect string
		actual string
	)

	backupCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	backupCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := backupCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'backupCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = ""
	actual = buffTmp.String() // resotre buffer
	assert.Equal(t, expect, actual,
		"Command 'build' with no parameters should produce an empty value.",
	)
}

func Test_backupCmd_Help(t *testing.T) {
	var (
		backupCmd = CreateBackupCmd()
		argsTmp   = []string{"--help"}
		buffTmp   = new(bytes.Buffer)

		expect string
		actual string
	)

	backupCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	backupCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := backupCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'backupCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = "Usage:"
	actual = buffTmp.String() // resotre buffer
	assert.Contains(t, actual, expect,
		"Command 'help' should show usage",
	)
}
