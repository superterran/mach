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

var restoreCmd = createRestoreCmd()

func createRestoreCmd() *cobra.Command {
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
	testMode = strings.HasSuffix(os.Args[0], ".test")

	viper.SetDefault("machine-s3-bucket", "mach-docker-machine-certificates")
	viper.SetDefault("machine-s3-region", "us-east-1")

}

func runRestore(cmd *cobra.Command, args []string) error {

	if len(args) == 1 {

		createTempDirectory()
		downloadFromS3(args[0])
		extractTarball(args[0])
		populateMachineDir(args[0])

		// defer os.RemoveAll(tmpDir)

	}

	return nil
}

func downloadFromS3(machine string) {

	var filename string = machine + ".tar.gz"
	var bucket = viper.GetString("machine-s3-bucket")
	var item string = filename
	file, err := os.Create(item)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("machine-s3-region"))},
	)
	if err != nil {
		log.Fatal(err)
	}

	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
}

func extractTarball(machine string) {

	gzipStream, err := os.Open(machine + ".tar.gz")
	if err != nil {
		fmt.Println("error")
	}

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		fmt.Println(err)

		log.Fatal("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(tmpDir + "/" + filepath.Base(header.Name))
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			log.Fatalf("ExtractTarGz: uknown type: %s in %s", header.Typeflag, header.Name)
		}

	}

}

func populateMachineDir(machine string) {

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(homedir)

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

	fmt.Println(dest)

	var src string = tmpDir + "/" + filepath.Base(dest)

	fmt.Println(src)

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
