/*
Copyright © 2021 Doug Hatcher <superterran@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

/* https://github.com/KEINOS/Hello-Cobra */

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildCmd(t *testing.T) {
	var (
		buildCmd = createBuildCmd()
		argsTmp  = []string{}
		buffTmp  = new(bytes.Buffer)

		expect string
		actual string
	)

	buildCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	buildCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := buildCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'buildCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = ""
	actual = buffTmp.String() // resotre buffer
	assert.Equal(t, expect, actual,
		"Command 'build' with no parameters should produce an empty value.",
	)
}

func Test_buildCmd_Help(t *testing.T) {
	var (
		buildCmd = createBuildCmd()
		argsTmp  = []string{"--help"}
		buffTmp  = new(bytes.Buffer)

		expect string
		actual string
	)

	buildCmd.SetOut(buffTmp)  // set output from os.Stdout -> buffTmp
	buildCmd.SetArgs(argsTmp) // set command args

	// Run `hello` command!
	if err := buildCmd.Execute(); err != nil {
		assert.FailNowf(t, "Failed to execute 'buildCmd.Execute()'.", "Error msg: %v", err)
	}

	expect = "Usage:"
	actual = buffTmp.String() // resotre buffer
	assert.Contains(t, actual, expect,
		"Command 'help' should show usage",
	)
}

func Test_BasicExampleBuild(t *testing.T) {
	var expect = "skipping image build"
	var actual = buildImage("images/example/Dockerfile", true)
	assert.Contains(t, actual, expect,
		"buildImage method should get to end, skipping image build due to testing state",
	)
}

func Test_BasicExamplePush(t *testing.T) {
	var expect = "skipping push due to testing"
	var actual = pushImage("example-varient", true)
	assert.Contains(t, actual, expect,
		"pushImage method should get to end, skipping push due to testing state",
	)
}
