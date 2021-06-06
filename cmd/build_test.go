package cmd

/* https://github.com/KEINOS/Hello-Cobra */

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildCmd(t *testing.T) {
	var (
		buildCmd = CreateBuildCmd()
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
		buildCmd = CreateBuildCmd()
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

func Test_BasicExamplePush(t *testing.T) {

	Nopush = true
	var expect = "skipping"
	var actual = pushImage("example-variant")
	assert.Contains(t, actual, expect,
		"pushImage method should get to end, skipping push due to testing state",
	)
}

func Test_BasicExamplePushNoPushFlag(t *testing.T) {
	var expect = "skipping push due to TestMode"
	var actual = pushImage("example-variant")
	assert.Contains(t, actual, expect,
		"pushImage method should get to end, skipping push due to testing state",
	)
}

func Test_BranchVariant(t *testing.T) {
	var expect = ""
	var actual = getBranchVariant()
	assert.Contains(t, actual, expect,
		"branch variant should come back as empty",
	)
}

func Test_GetTag(t *testing.T) {
	var expect = "superterran/mach:example"
	var actual = getTag("images/example/Dockerfile")
	assert.Contains(t, actual, expect,
		"tag should come back as superterran/mach:example",
	)
}

func Test_GetTagNoRegistry(t *testing.T) {
	DockerRegistry = ""
	var expect = "example"
	var actual = getTag("images/example/Dockerfile")
	assert.Contains(t, actual, expect,
		"tag should come back as example",
	)
}

func Test_GetTagWithVariant(t *testing.T) {
	DockerRegistry = "superterran/mach"
	var expect = "superterran/mach:example-test"
	var actual = getTag("images/example/Dockerfile-test")
	assert.Contains(t, actual, expect,
		"tag should come back as superterran/mach:example-test",
	)
}

func Test_dockerLogStream(t *testing.T) {
	var expect = "Successfully built 6dbb9cc54074"
	var actual = dockerLog("{\"stream\":\"Successfully built 6dbb9cc54074\n\"}")
	assert.Contains(t, actual, expect,
		"dockerLog method only contains json body",
	)
}

func Test_dockerLogStatus(t *testing.T) {
	var expect = "Successfully built 6dbb9cc54074"
	var actual = dockerLog("{\"status\":\"Successfully built 6dbb9cc54074\n\"}")
	assert.Contains(t, actual, expect,
		"dockerLog method only contains json body",
	)
}

func Test_dockerLogStrangeMessageInFull(t *testing.T) {
	var expect = "blah"
	var actual = dockerLog("blah")
	assert.Contains(t, actual, expect,
		"dockerLog method returns strange data as-is",
	)
}

func Test_buildCmdOneArgNone(t *testing.T) {

	var actual = MainBuildFlow([]string{""})

	if actual != nil {
		assert.FailNowf(t, "returned not nil.", "Error msg: %v", actual)
	}
}

func Test_buildCmdOneArgOutputOnly(t *testing.T) {

	OutputOnly = true

	var actual = MainBuildFlow([]string{"example"})

	if actual != nil {
		assert.FailNowf(t, "returned not nil.", "Error msg: %v", actual)
	}
}

func Test_buildCmdOneArgOutputOnlyWithTag(t *testing.T) {

	OutputOnly = true

	var actual = MainBuildFlow([]string{"example:go"})

	if actual != nil {
		assert.FailNowf(t, "returned not nil.", "Error msg: %v", actual)
	}
}

func Test_buildCmdOneArgWithImagesOutputOnlyWithTag(t *testing.T) {

	OutputOnly = true

	var actual = MainBuildFlow([]string{"example:go"})

	if actual != nil {
		assert.FailNowf(t, "returned not nil.", "Error msg: %v", actual)
	}
}

func Test_buildCmdOneArgWithImagesOutputOnlyWithTagReal(t *testing.T) {

	OutputOnly = false
	TestMode = false

	var actual = buildImage("../examples/images/example/Dockerfile")

	if actual != "superterran/mach:example" {
		assert.FailNowf(t, "mach tag returned as expected, %s", actual)
	}
}

func Test_buildCmdOneArgWithImagesOutputOnlyWithTagRealTemplate(t *testing.T) {

	OutputOnly = false
	TestMode = false

	var actual = buildImage("../examples/images/example/Dockerfile-template.tpl")

	if actual != "superterran/mach:example-template" {
		assert.FailNowf(t, "mach tag returned as expected, %s", actual)
	}
}

func Test_buildCmdOneArgWithImagesOutputOnlyWithTagRealOutputOnly(t *testing.T) {

	OutputOnly = true
	TestMode = false

	var actual = buildImage("../examples/images/example/Dockerfile")

	if actual != "superterran/mach:example" {
		assert.FailNowf(t, "mach tag returned as expected, %s", actual)
	}
}

func Test_buildCmdOneArgWithImagesOutputOnlyWithTagRealTemplateOutputOnly(t *testing.T) {

	OutputOnly = true
	TestMode = false

	var actual = buildImage("../examples/images/example/Dockerfile-template.tpl")

	if actual != "superterran/mach:example-template" {
		assert.FailNowf(t, "mach tag returned as expected, %s", actual)
	}
}

func Test_buildCmdOneArgWithImagesAndPush(t *testing.T) {

	OutputOnly = false
	TestMode = false
	Nopush = false

	// DockerUser = os.Getenv("MACH_DOCKER_USER")
	// DockerPassword = os.Getenv("MACH_DOCKER_PASS")

	var tag = buildImage("../examples/images/example/Dockerfile-template.tpl")
	var actual = pushImage(tag)

	if actual != "push complete" {
		assert.FailNowf(t, "mach tag returned as expected, %s", actual)
	}
}
