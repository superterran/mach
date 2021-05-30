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
	MachineS3Bucket = viper.GetString("machine-s3-bucket")

	viper.SetDefault("machine-s3-region", MachineS3Region)
	MachineS3Region = viper.GetString("machine-s3-region")

	restoreCmd.Flags().BoolP("keep-tarball", "k", KeepTarball, "keeps the tarball in working directory after upload")
	KeepTarball, _ := restoreCmd.Flags().GetBool("keep-tarball")
	if KeepTarball {
		fmt.Println("--keep-tarball set")
	}

}

func runRestore(cmd *cobra.Command, args []string) error {

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

	gzipStream, err := os.Open(machine + ".tar.gz")
	if err != nil {
		fmt.Println("error")
	}

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		fmt.Println(err)

		log.Fatal("tarball: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("tarball: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				log.Fatalf("tarball: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(tmpDir + "/" + filepath.Base(header.Name))
			if err != nil {
				log.Fatalf("tarball: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("tarball: Copy() failed: %s", err.Error())
			}
			outFile.Close()
		}

	}

}

func populateMachineDir(machine string) {

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	var machinedir = homedir + "/.docker/machine/machines/" + machine + "/"
	var certsdir = homedir + "/.docker/machine/certs/"

	os.Mkdir(machinedir, 0755)
	os.Mkdir(certsdir, 0755)

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

}

func replaceInMachineFile(file string, new string, old string) {

	input, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
	}

	var output string = strings.ReplaceAll(string(input), old, new)

	err = ioutil.WriteFile(file, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
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

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	if err != nil {
		return 0, err
	}
	return nBytes, err
}
