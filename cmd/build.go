/*
The build command generates docker images, using templates, and pushes them to registries. This provides a useful
method to rapidly build images with variants, and manage them in a git repository. Dockerfiles can be made
supporting includes, conditionals, loops, etc.

Each image should get it's own directory in the images directory, by default the current working dir. If mach
build is ran with no arguments, it will try to build every image discovered.

- images/<image_name>/Dockerfile is the default image to be built, this is a proper Dockerfile and should not
  use the templating system.

- images/<image_name>/Dockerfile-<variant> The variant can be used to build alternate images, when pushed to
  the registry the variant is appended to the image name.

- images/<image_name>/Dockerfile[-<variant>].tpl using the .tpl will process through the templating engine. This
  allows for including partial templates.
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

var nopush bool = false

var lastOutput = "begin"

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

	/* no-push flag prevents pushing to the docker registry, good for testing locally */
	buildCmd.Flags().BoolP("no-push", "n", false, "Do not push to registry")

	/* output-only does not build the image, just outputs the template to stdout */
	buildCmd.Flags().BoolP("output-only", "o", false, "send output to stdout, do not build")

	/* first-only will stop after processing the first image if more than one image is matched. Good to pair with output-only
	   if you intend to pipe output to a file */
	buildCmd.Flags().BoolP("first-only", "f", false, "breaks after processing first image if more than one")

	// buildImageDirname tells the tool which directory to itereate through to find Dockerfiles. defaults the present working
	// directory, but a good practice is to mint a .mach.yaml and set this to `images` or the like when building an IaC repo.
	viper.SetDefault("buildImageDirname", ".")

	// defaultGitBranch allows for setting which branch does not add a branch variant to the tag. Default to main, consider
	// changing your branch name before chaning this default.
	viper.SetDefault("defaultGitBranch", "main")

	//docker_host is the URL for the docker registry, we default to the offical registry, but this can be changed in config
	// to any registry you like.
	viper.SetDefault("docker_host", "https://index.docker.io/v1/")

	if testMode {
		// docker_registry is the package name inside the docker registry. @todo think through a better name
		viper.SetDefault("docker_registry", "superterran/mach")
	}
}

/* runBuild is the main flow for the build command. If no arguments are present, it will each through
the images directory and attempt to build every file matching the pattern `Dockerfile*`. If arguements are passed
it will attempt to match the strings with dockerfiles and build those only. This method sets several flags, outputOnly
is suitable for leveraging the stdout which provides the contents of a templated dockerfile. */
func runBuild(cmd *cobra.Command, args []string) error {

	outputonly, _ := cmd.Flags().GetBool("output-only")
	if outputonly {
		testMode = true
		outputOnly = true
	}

	fnopush, _ := cmd.Flags().GetBool("no-push")
	nopush = fnopush

	firstOnly, _ := cmd.Flags().GetBool("first-only")

	if len(args) < 1 {
		matches, _ := filepath.Glob(viper.GetString("buildImageDirname") + "/**/Dockerfile*")
		for _, match := range matches {
			var mach_tag string = buildImage(match)
			if !nopush || outputonly {
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

// generateTemplate grabs the docker tpl file, and any tpl files in the `includes` sub directory with the
// docerfile, and runs them through a templater to produce the output for a dockerfile to be built.
// this method uses the `html/template` package https://golang.org/pkg/html/template/ so this should be
// fairly flexible. I intentionally haven't introduced outside variables to the templating engine i.e.
// host system environment variables. This may come in time, but Dockerfiles should not contain secrets so
// I'm not sure if this is a good feature to introduce.
func generateTemplate(wr io.Writer, filename string) {

	tpl, err := template.ParseGlob(filename)
	if err != nil {
		panic(err)
	}

	tpl.ParseGlob(filepath.Dir(filename) + "/includes/*.tpl")
	tpl.Execute(wr, filepath.Base(filename))

}

// getTag determines the string used for the docker image tag that gets referenced in the registry.
// This is typically just the directory name the dockerfile resides in, but can be modified by either
// using a variant in the filename (detailed below) or by using a git branch other than the default
func getTag(filename string) string {

	var tag string = filepath.Base(filepath.Dir(filename)) + getVariant(filename)

	if viper.GetString("docker_registry") != "" {
		return viper.GetString("docker_registry") + ":" + tag
	} else {
		nopush = true
		return tag
	}

}

// getVariant determines the additional string to append to the docker image tag. This is determined
// by the filename of the docker file being read, `Dockerfile` will not get any variant, but files like
// `Dockerfile-variant` will be tagged as `<dirname>-<variant>`.
func getVariant(filename string) string {

	var variant string = ""

	if strings.Contains(filepath.Base(filename), "-") {
		variant = "-" + strings.Split(filepath.Base(filename), "-")[1]
	}

	return strings.Replace(variant, ".tpl", "", 1) + getBranchVariant()

}

// getBranchVariant will return a string that can appended to the variant of a tag, this function is called
// by getVariant. This will not produce output if on the default branch-name, otherwise it will return
// a string with the branch.
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

// buildImage probably does too much, but it creates a tarball with a templatized dockerfile, and
// everything in it's directory for context, and it builds the image.
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

	e := os.Remove(DockerFilename)
	if e != nil {
		log.Fatal(e)
	}

	return mach_tag

}

// pushImage takes the current tag and pushes it to the configured registry. This fuction short-circuits
// if testMode or nopush are true.
func pushImage(mach_tag string) string {

	if testMode {
		return "skipping push due to testMode"
	}

	if nopush {
		return "in no push mode, skipping"
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

// dockerLog is a logging method that tries to handle the output produced by the docker daemon.
// it needs a bit of work.
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

		}
	}

	return msg
}
