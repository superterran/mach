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

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/moby/term"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = createBuildCmd()

var testMode = false

var outputOnly = false

var firstOnly = false

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func createBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [docker-image[:tag]]",
		Short: "Builds a directory of docker images and pushes them to a registry",
		Long: `This allows you to maintain a directory of docker images, with templating,
	and use this to populate a docker registry. `,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(cmd, args)
		},
	}
	return cmd
}

func init() {

	rootCmd.AddCommand(buildCmd)

	testMode = strings.HasSuffix(os.Args[0], ".test")

	buildCmd.Flags().BoolP("no-push", "n", false, "Do not push to registry")
	buildCmd.Flags().BoolP("output-only", "o", false, "send output to stdout, do not build")
	buildCmd.Flags().BoolP("first-only", "f", false, "breaks after processing first image if more than one")

	viper.SetDefault("buildImageDirname", "./images")
	viper.SetDefault("defaultGitBranch", "main")
	viper.SetDefault("docker_host", "https://index.docker.io/v1/")

	if testMode {
		viper.SetDefault("docker_registry", "superterran/mach")
	}
}

func runBuild(cmd *cobra.Command, args []string) error {

	outputonly, _ := cmd.Flags().GetBool("output-only")
	if outputonly {
		testMode = true
		outputOnly = true
	}

	fnopush, _ := cmd.Flags().GetBool("no-push")

	firstOnly, _ := cmd.Flags().GetBool("first-only")

	if len(args) < 1 {
		matches, _ := filepath.Glob(viper.GetString("buildImageDirname") + "/**/Dockerfile*")
		for _, match := range matches {
			var mach_tag string = buildImage(match)
			if !fnopush || outputonly {
				pushImage(mach_tag)
			}

			if firstOnly {
				break
			}
		}
	} else {
		for _, arg := range args {

			var image string = arg
			var variant string

			if strings.Contains(arg, ":") {
				image = strings.Split(arg, ":")[0]
				variant = "-" + strings.Split(arg, ":")[1]
			}

			matches, _ := filepath.Glob(viper.GetString("buildImageDirname") + "/" + image + "/Dockerfile" + variant + "*")
			for _, match := range matches {
				var mach_tag string = buildImage(match)

				if !fnopush || outputonly {
					pushImage(mach_tag)
				}

				if firstOnly {
					break
				}
			}
		}
	}

	return nil
}

func generateTemplate(wr io.Writer, filename string) {

	tpl, err := template.ParseGlob(filename)
	if err != nil {
		panic(err)
	}

	tpl.ParseGlob(filepath.Dir(filename) + "/includes/*.tpl")
	tpl.Execute(wr, filepath.Base(filename))

}

func getBranchVariant() string {

	var branch = ""
	var variant string = ""

	repo, err := git.PlainOpen(".")
	if err != nil {
		branch = "origin/refs/main"
	} else {

		head, err := repo.Head()
		if err != nil {
			log.Fatal(err)
		}

		branch = head.String()
	}

	if strings.Contains(branch, "/") {
		var variant_branch string = strings.Split(branch, "/")[2]
		if variant_branch != viper.GetString("defaultGitBranch") {
			variant = "-" + variant_branch
		}
	}

	return variant
}

func getVariant(filename string) string {

	var variant string = ""

	if strings.Contains(filepath.Base(filename), "-") {
		variant = "-" + strings.Split(filepath.Base(filename), "-")[1]
	}

	return strings.Replace(variant, ".tpl", "", 1) + getBranchVariant()

}

func getTag(filename string) string {
	return viper.GetString("docker_registry") + ":" + filepath.Base(filepath.Dir(filename)) + getVariant(filename)
}

func buildImage(filename string) string {

	var mach_tag = getTag(filename)

	if !outputOnly {
		color.HiYellow("Building image with tag " + mach_tag)
	}

	var DockerFilename string = filepath.Dir(filename) + "/." + filepath.Base(filename) + ".generated"

	if outputOnly || testMode {

		generateTemplate(os.Stdout, filename)
		return mach_tag

	} else {
		f, err := os.Create(DockerFilename)
		if err != nil {
			log.Fatal(err)
		}

		generateTemplate(f, filename)

		f.Close()
	}

	if !testMode {
		tar, err := archive.TarWithOptions(filepath.Dir(DockerFilename), &archive.TarOptions{})
		if err != nil {
			log.Fatal(err)
		}

		opts := types.ImageBuildOptions{
			Dockerfile: filepath.Base(DockerFilename),
			Remove:     true,
			Tags:       []string{mach_tag},
		}

		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatal(err)
		}

		res, err := cli.ImageBuild(ctx, tar, opts)

		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {

			var lastLine = scanner.Text()

			errLine := &ErrorLine{}
			json.Unmarshal([]byte(lastLine), errLine)
			if errLine.Error != "" {
				log.Fatal(color.RedString(errLine.Error))
			} else {
				dockerLog(scanner.Text())
			}
		}
	}

	return mach_tag

}

func pushImage(mach_tag string) string {

	if testMode {
		return "skipping push due to testMode"
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}

	var authConfig = types.AuthConfig{
		Username:      viper.GetString("docker_user"),
		Password:      viper.GetString("docker_pass"),
		ServerAddress: viper.GetString("docker_host"),
	}
	authConfigBytes, _ := json.Marshal(authConfig)
	authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	tag := mach_tag

	opts := types.ImagePushOptions{RegistryAuth: authConfigEncoded}
	rd, err := cli.ImagePush(ctx, tag, opts)

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(rd, os.Stderr, termFd, isTerm, nil)

	if err != nil {
		log.Fatal(err)
	}

	if rd == nil {
		log.Fatal(rd)
	}

	defer rd.Close()

	return "push complete"
}

func dockerLog(msg string) string {

	var result map[string]interface{}
	json.Unmarshal([]byte(msg), &result)

	for key, value := range result {

		switch msgtype := key; msgtype {

		case "status":
			color.Yellow(value.(string))
			return value.(string)
		case "stream":
			color.Blue(value.(string))
			return value.(string)
		case "aux":
		case "errorDetail":
		default:
			return value.(string)
		}
	}

	return msg
}
