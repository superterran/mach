// Cmd restore copies docker-machine certs and configurations from S3 and applies them to the host
package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var restoreCmd = CreateRestoreCmd()

func CreateRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore <docker-machine>",
		Short: "Restores a docker-machine tarball from S3 to this machine for use.",
		Long: `This allows you to restore the docker-machine certs bundle from an S3 bucket, 
paired with the backup command, this will let you transfer docker-machines to and from
systems using the AWS API. Will require progamtic credetials with permissions to dowmload from S3.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestore(cmd, args)
		},
	}
	return cmd
}

func init() {

	rootCmd.AddCommand(restoreCmd)

	viper.SetDefault("machine-s3-bucket", MachineS3Bucket)

	viper.SetDefault("machine-s3-region", MachineS3Region)

	restoreCmd.Flags().BoolP("keep-tarball", "k", KeepTarball, "keeps the tarball in working directory after upload")

}

func runRestore(cmd *cobra.Command, args []string) error {

	MachineS3Bucket = viper.GetString("machine-s3-bucket")

	MachineS3Region = viper.GetString("machine-s3-region")

	KeepTarball, _ := cmd.Flags().GetBool("keep-tarball")
	if KeepTarball {
		fmt.Println("--keep-tarball set")
	}

	if len(args) == 1 {

		createTempDirectory()
		downloadFromS3(args[0])
		extractTarball(args[0])
		populateMachineDir(args[0])

		defer os.RemoveAll(tmpDir)

		if !KeepTarball {
			removeMachineArchive(args[0])
		}

		fmt.Println(args[0] + " restore complete from " + MachineS3Bucket + " bucket")

	}

	return nil
}

func downloadFromS3(machine string) {

	var filename string = machine + ".tar.gz"
	var item string = filename
	file, err := os.Create(item)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(MachineS3Region)},
	)
	if err != nil {
		log.Fatal(err)
	}

	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(MachineS3Bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		fmt.Println("Failed to download", file.Name(), numBytes, "bytes")
		log.Fatal(err)
	}

}

func extractTarball(machine string) {

	var path = ""

	if TestMode {
		path = "../examples/"
	}
	gzipStream, _ := os.Open(path + machine + ".tar.gz")

	uncompressedStream, _ := gzip.NewReader(gzipStream)

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		switch header.Typeflag {
		// case tar.TypeDir:
		// 	os.Mkdir(header.Name, 0755)
		case tar.TypeReg:
			outFile, _ := os.Create(tmpDir + "/" + filepath.Base(header.Name))
			io.Copy(outFile, tarReader)
			outFile.Close()
		}

	}

}

func populateMachineDir(machine string) bool {

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if TestMode {
		homedir = tmpDir
	}

	var machinedir = homedir + "/.docker/machine/machines/" + machine + "/"
	var certsdir = homedir + "/.docker/machine/certs/"

	os.MkdirAll(machinedir, 0755)
	os.MkdirAll(certsdir, 0755)

	copyTo(machinedir + "ca.pem")
	copyTo(machinedir + "cert.pem")

	copyTo(machinedir + "config.json.template")
	copyTo(machinedir + "key.pem")
	copyTo(machinedir + "server-key.pem")
	copyTo(machinedir + "server.pem")
	copyTo(certsdir + "ca-key.pem")
	copyTo(certsdir + "ca.pem")
	copyTo(certsdir + "cert.pem")
	copyTo(certsdir + "key.pem")

	var config = machinedir + "config.json"

	copyTo(config)
	replaceInMachineFile(config, homedir, "${TEMPLATE_HOME_DIR}")
	replaceInMachineFile(config, machine, "${TEMPLATE_MACHINE_NAME}")

	// makes compatible with bash script
	replaceInMachineFile(config, homedir, "___HOME_DIR___")
	replaceInMachineFile(config, machine, "___MACHINE_NAME___")

	return true

}

func replaceInMachineFile(file string, new string, old string) {

	input, _ := ioutil.ReadFile(file)

	var output string = strings.ReplaceAll(string(input), old, new)

	ioutil.WriteFile(file, []byte(output), 0644)

}

func copyTo(dest string) (int64, error) {

	var src string = tmpDir + "/" + filepath.Base(dest)

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", dest)
	}

	source, _ := os.Open(src)

	defer source.Close()

	destination, _ := os.Create(dest)

	defer destination.Close()
	nBytes, err := io.Copy(destination, source)

	return nBytes, err
}

func removeMachineArchive(machine string) {
	e := os.Remove(machine + ".tar.gz")
	if e != nil {
		log.Fatal(e)
	}
}
