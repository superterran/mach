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
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
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

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds a directory of docker images and pushes them to a registry",
	Long: `This allows you to maintain a directory of docker images, with templating,
	and use this to populate a docker registry. `,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(viper.Get("example"))

		fmt.Println(viper.Get("registry"))

		matches, _ := filepath.Glob(viper.GetString("buildImageDirname") + "/**/Dockerfile*")
		for _, match := range matches {

			buildImage(match)

		}
	},
}

func buildImage(filename string) {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	tar, err := archive.TarWithOptions(filename, &archive.TarOptions{})
	if err != nil {
		log.Fatal(err)
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Remove:     true,
	}

	res, err := cli.ImageBuild(ctx, tar, opts)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {

		var lastLine = scanner.Text()
		fmt.Println(scanner.Text())

		errLine := &ErrorLine{}
		json.Unmarshal([]byte(lastLine), errLine)
		if errLine.Error != "" {
			log.Fatal(errLine.Error)
		}
	}
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	buildCmd.Flags().BoolP("no-push", "n", false, "Do not push to registry")

	viper.SetDefault("buildImageDirname", "./images")

}
