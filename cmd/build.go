// Cmd build generates docker images, using templates, and pushes them to a registry
package cmd

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

var buildCmd = CreateBuildCmd()

// FirstOnly will stop the build loop after the first image is found, useful for output only
var FirstOnly bool = false

// Nopush builds the image, but does not push to a registry. Set with `--no-push` or `-n`
var Nopush bool = false

// BuildImageDirname tells the tool which directory to itereate through to find Dockerfiles. defaults the present working
// directory, but a good practice is to mint a .mach.yaml and set this to `images` or the like when building an IaC repo.
var BuildImageDirname string = "."

// DefaultGitBranch allows for setting which branch does not add a branch variant to the tag. Default to main, consider
// changing your branch name before chaning this default.
var DefaultGitBranch string = "main"

// DockerHost is the URL for the docker registry, we default to the offical registry, but this can be changed in config
// to any registry you like.
var DockerHost string = "https://index.docker.io/v1/"

// DockerRegistry is the package name inside the docker registry. @todo think through a better name
var DockerRegistry string = "superterran/mach"

// DockerUser is the registry username
var DockerUser string = ""

// DockerPassword is the registry password
var DockerPassword string = ""

// Verbose removes the terminal formatting for builds, displaying the entire output to the user
var Verbose bool = false

// builds docker images without cache
var NoCache bool = false

type errorLine struct {
	Error       string      `json:"error"`
	errorDetail errorDetail `json:"errorDetail"`
}

type errorDetail struct {
	Message string `json:"message"`
}

func CreateBuildCmd() *cobra.Command {
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

	buildCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is loaded from working dir)")

	buildCmd.Flags().BoolP("no-push", "n", Nopush, "Do not push to registry")

	buildCmd.Flags().BoolP("no-cache", "c", NoCache, "no build cache")

	buildCmd.Flags().BoolP("output-only", "o", OutputOnly, "send output to stdout, do not build")

	buildCmd.Flags().BoolP("first-only", "f", FirstOnly, "stop the build loop after the first image is found")

	buildCmd.Flags().BoolP("verbose", "v", Verbose, "show entire build output")

	buildCmd.Flags().StringVar(&BuildImageDirname, "build-image-dir-name", BuildImageDirname, "build Image directory")
	viper.SetDefault("BuildImageDirname", BuildImageDirname)
	viper.BindPFlag("BuildImageDirname", buildCmd.Flags().Lookup("build-image-dir-name"))

	buildCmd.Flags().StringVar(&DefaultGitBranch, "default-git-branch", DefaultGitBranch, "default git branch")
	viper.SetDefault("defaultGitBranch", DefaultGitBranch)
	viper.BindPFlag("defaultGitBranch", buildCmd.Flags().Lookup("default-git-branch"))

	buildCmd.Flags().StringVar(&DockerRegistry, "docker-registry", DockerRegistry, "docker registry")
	viper.SetDefault("docker_registry", DockerRegistry)
	viper.BindPFlag("docker_registry", buildCmd.Flags().Lookup("docker-registry"))

	buildCmd.Flags().StringVar(&DockerHost, "docker-host", DockerHost, "docker registry hostname")
	viper.SetDefault("docker_host", DockerHost)
	viper.BindPFlag("docker_host", buildCmd.Flags().Lookup("docker-host"))

	buildCmd.Flags().String("docker-user", DockerUser, "docker registry username")
	viper.SetDefault("docker_user", DockerUser)
	viper.BindPFlag("docker_user", buildCmd.Flags().Lookup("docker-user"))

	buildCmd.Flags().StringVar(&DockerPassword, "docker-pass", DockerPassword, "docker registry password")
	viper.SetDefault("docker_pass", DockerPassword)
	viper.BindPFlag("docker_pass", buildCmd.Flags().Lookup("docker-pass"))

}

// runBuild is the main flow for the build command. If no arguments are present, it will each through
// the images directory and attempt to build every file matching the pattern `Dockerfile*`. If arguements are passed
// it will attempt to match the strings with dockerfiles and build those only. This method sets several flags,
// OutputOnly is suitable for leveraging the stdout which provides the contents of a templated dockerfile.
func runBuild(cmd *cobra.Command, args []string) error {

	Nopush, _ = cmd.Flags().GetBool("no-push")

	OutputOnly, _ = cmd.Flags().GetBool("output-only")

	FirstOnly, _ = cmd.Flags().GetBool("first-only")

	BuildImageDirname = viper.GetString("BuildImageDirname")

	DockerHost = viper.GetString("docker_host")

	DockerUser = viper.GetString("docker_user")

	DockerPassword = viper.GetString("docker_pass")

	DockerRegistry = viper.GetString("docker_registry")

	return MainBuildFlow(args)
}

// MainBuildFlow will run builds against an array of arguments, if no arguments are supplied
// it will iterate through the build directory
func MainBuildFlow(args []string) error {
	if OutputOnly {
		TestMode = true
		OutputOnly = true
	}

	if len(args) < 1 {
		matches, _ := filepath.Glob(BuildImageDirname + "/**/Dockerfile*")
		for _, match := range matches {
			var mach_tag string = buildImage(match)
			if !Nopush || OutputOnly {
				pushImage(mach_tag)
			}

			if FirstOnly {
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

			matches, _ := filepath.Glob(BuildImageDirname + "/" + image + "/Dockerfile" + variant + "*")
			for _, match := range matches {
				var mach_tag string = buildImage(match)

				if !Nopush || OutputOnly {
					pushImage(mach_tag)
				}

				if FirstOnly {
					break
				}
			}
		}
	}

	return nil
}

// generateDockerfileTemplate grabs the docker tpl file, and any tpl files in the `includes` sub directory with the
// docerfile, and runs them through a templater to produce the output for a dockerfile to be built.
// this method uses the `html/template` package https://golang.org/pkg/html/template/ so this should be
// fairly flexible. I intentionally haven't introduced outside variables to the templating engine i.e.
// host system environment variables. This may come in time, but Dockerfiles should not contain secrets so
// I'm not sure if this is a good feature to introduce.
func generateDockerfileTemplate(wr io.Writer, filename string) {

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

	if DockerRegistry != "" {
		return DockerRegistry + ":" + tag
	} else {
		Nopush = true
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
		if variant_branch != DefaultGitBranch {
			variant = "-" + variant_branch
		}
	}

	return variant
}

// buildImage probably does too much, but it creates a tarball with a templatized dockerfile, and
// everything in it's directory for context, and it builds the image.
func buildImage(filename string) string {

	var mach_tag = getTag(filename)

	if !OutputOnly {
		color.HiYellow("Building image with tag " + mach_tag)
	}

	if OutputOnly || TestMode {

		generateDockerfileTemplate(os.Stdout, filename)
		return mach_tag

	}

	fmt.Print("\033[s")

	var DockerFilename string = filepath.Dir(filename) + "/." + filepath.Base(filename) + ".generated"

	f, _ := os.Create(DockerFilename)
	generateDockerfileTemplate(f, filename)

	f.Close()

	tar, _ := archive.TarWithOptions(filepath.Dir(DockerFilename), &archive.TarOptions{})

	opts := types.ImageBuildOptions{
		Dockerfile: filepath.Base(DockerFilename),
		Remove:     true,
		Tags:       []string{mach_tag},
		NoCache:    NoCache,
	}

	ctx := context.Background()
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	res, _ := cli.ImageBuild(ctx, tar, opts)

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {

		var lastLine = scanner.Text()

		errLine := &errorLine{}
		json.Unmarshal([]byte(lastLine), errLine)
		if errLine.Error != "" {
			log.Fatal(color.RedString(errLine.Error))
		} else {
			dockerLog(scanner.Text())
		}
	}

	e := os.Remove(DockerFilename)
	if e != nil {
		log.Fatal(e)
	}

	return mach_tag

}

// pushImage takes the current tag and pushes it to the configured registry. This fuction short-circuits
// if TestMode or Nopush are true.
func pushImage(mach_tag string) string {

	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	var authConfig = types.AuthConfig{
		Username:      DockerUser,
		Password:      DockerPassword,
		ServerAddress: DockerHost,
	}
	authConfigBytes, _ := json.Marshal(authConfig)
	authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)

	tag := mach_tag

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*60)
	defer cancel()

	opts := types.ImagePushOptions{RegistryAuth: authConfigEncoded}

	termFd, isTerm := term.GetFdInfo(os.Stderr)

	if Nopush || TestMode {
		return "skipping push due to TestMode"
	}

	rd, err := cli.ImagePush(ctx, tag, opts)
	err = jsonmessage.DisplayJSONMessagesStream(rd, os.Stderr, termFd, isTerm, nil)
	if err != nil {
		if !TestMode {
			log.Fatal(err)
		}
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
			color.Yellow(value.(string) + "\n\n")
			return value.(string)
		case "stream":
			scanner := bufio.NewScanner(strings.NewReader(value.(string)))
			for scanner.Scan() {
				if !Verbose {
					fmt.Print("\033[u\033[2K")
				}

				color.Blue(scanner.Text())
			}

			return value.(string)
		case "errorDetail":
			color.Red(value.(string))
		default:

		}
	}

	return msg
}
