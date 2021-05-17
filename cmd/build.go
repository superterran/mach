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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

var buildCmd = &cobra.Command{
	Use:   "build [docker-image[:tag]]",
	Short: "Builds a directory of docker images and pushes them to a registry",
	Long: `This allows you to maintain a directory of docker images, with templating,
and use this to populate a docker registry. `,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Println("building all")
			matches, _ := filepath.Glob(viper.GetString("buildImageDirname") + "/**/Dockerfile*")
			for _, match := range matches {
				buildImage(match, cmd)
			}

		} else {

			fmt.Println("building args that match")
			for _, arg := range args {

				var image string = arg
				var variant string

				if strings.Contains(arg, ":") {
					image = strings.Split(arg, ":")[0]
					variant = "-" + strings.Split(arg, ":")[1]
				}

				matches, _ := filepath.Glob(viper.GetString("buildImageDirname") + "/" + image + "/Dockerfile" + variant + "*")
				for _, match := range matches {
					buildImage(match, cmd)
				}
			}
		}
	},
}

func buildImage(filename string, cmd *cobra.Command) {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	var variant string = ""

	if strings.Contains(filepath.Base(filename), "-") {
		variant = "-" + strings.Split(filepath.Base(filename), "-")[1]
	}

	variant = strings.Replace(variant, ".tpl", "", 1)

	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatal(err)
	}

	head, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(head.String(), "/") {
		var variant_branch string = strings.Split(head.String(), "/")[2]
		if variant_branch != viper.GetString("defaultGitBranch") {
			variant = "-" + variant_branch
		}
	}

	var mach_tag = viper.GetString("docker_registry") + ":" + filepath.Base(filepath.Dir(filename)) + variant

	fmt.Println("Building image with tag " + mach_tag)

	var DockerFilename string = filepath.Base(filename)

	if filepath.Ext(filename) == ".tpl" {
		tpl, err := template.ParseGlob(filename)
		if err != nil {
			panic(err)
		}

		tpl.ParseGlob(filepath.Dir(filename) + "/includes/*.tpl")

		DockerFilename = "." + strings.TrimSuffix(filepath.Base(filename), ".tpl") + ".generated"

		f, err := os.Create(filepath.Dir(filename) + "/" + DockerFilename)

		if err != nil {
			panic(err)
		}

		err = tpl.Execute(f, filepath.Base(filename))
		if err != nil {
			log.Print("execute: ", err)
			return
		}

		f.Close()
	}

	tar, err := archive.TarWithOptions(filepath.Dir(filename)+"/", &archive.TarOptions{})
	if err != nil {
		log.Fatal(err)
	}

	opts := types.ImageBuildOptions{
		Dockerfile: DockerFilename,
		Remove:     true,
		Tags:       []string{mach_tag},
	}

	res, err := cli.ImageBuild(ctx, tar, opts)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {

		var lastLine = scanner.Text()
		dockerLog(scanner.Text())

		errLine := &ErrorLine{}
		json.Unmarshal([]byte(lastLine), errLine)
		if errLine.Error != "" {
			dockerLog(errLine.Error)
			log.Fatal(errLine.Error)
		}
	}

	fnopush, _ := cmd.Flags().GetBool("no-push")
	if !fnopush {
		pushImage(mach_tag)
	}
}

func pushImage(mach_tag string) {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	fmt.Println("attempting to push image " + mach_tag)

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

	// io.Copy(os.Stdout, rd)

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(rd, os.Stderr, termFd, isTerm, nil)

	if err != nil {
		log.Fatal(err)
	}

	if rd == nil {
		log.Fatal(rd)
	}

	defer rd.Close()

}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolP("no-push", "n", false, "Do not push to registry")

	viper.SetDefault("buildImageDirname", "./images")
	viper.SetDefault("defaultGitBranch", "main")
	viper.SetDefault("docker_host", "https://index.docker.io/v1/")

}

func dockerLog(msg string) {
	var result map[string]interface{}
	json.Unmarshal([]byte(msg), &result)

	for key, value := range result {
		// Each value is an interface{} type, that is type asserted as a string
		switch msgtype := key; msgtype {
		case "status":
			color.Yellow(value.(string))
		case "stream":
			color.Blue(value.(string))
		case "aux":
			color.Green(msg)
		default:
			color.White(msg)
		}

	}
}
