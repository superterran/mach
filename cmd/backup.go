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

var backupCmd = createBackupCmd()

func createBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup <docker-machine>",
		Short: "Takes a working docker-machine entry and stores it to an S3 bucket",
		Long: `This allows you to store the docker-machine certs bundle in an S3 bucket, 
paired with restore command, this will let you transfer docker-machines to and from
systems using the AWS API. Will require progamtic credetials with permissions to upload
to S3.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackup(cmd, args)
		},
	}
	return cmd
}

func init() {

	rootCmd.AddCommand(backupCmd)
	testMode = strings.HasSuffix(os.Args[0], ".test")

	viper.SetDefault("machine-s3-bucket", "mach-docker-machine-certificates")
	viper.SetDefault("machine-s3-region", "us-east-1")

	backupCmd.Flags().BoolP("create", "c", false, "create the bucket before attempting backup")

}

func runBackup(cmd *cobra.Command, args []string) error {

	createFirst, _ := cmd.Flags().GetBool("create")

	if createFirst {
		createBucket()
	}

	if len(args) == 1 {
		createTempDirectory()
		populateTempDir(args[0])
		createMachineTarball(args[0])
		uploadFileToBucket(args[0])

		defer os.RemoveAll(tmpDir)
	}

	return nil
}

func populateTempDir(machine string) {

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(homedir)

	var machinedir = homedir + "/.docker/machine/machines/" + machine + "/"
	var certsdir = homedir + ".docker/machine/certs/"

	copy(machinedir + "ca.pem")
	copy(machinedir + "cert.pem")

	copy(machinedir + "config.json.template")
	copy(machinedir + "key.pem")
	copy(machinedir + "server-key.pem")
	copy(machinedir + "server.pem")
	copy(certsdir + "ca-key.pem")
	copy(certsdir + "ca.pem")
	copy(certsdir + "cert.pem")
	copy(certsdir + "key.pem")

	var config = machinedir + "config.json"

	copy(config)
	replaceInTempFile(config, homedir, "${TEMPLATE_HOME_DIR}")
	replaceInTempFile(config, machine, "${TEMPLATE_MACHINE_NAME}")

}

func replaceInTempFile(file string, old string, new string) {

	var tempfile string = tmpDir + "/" + filepath.Base(file)

	input, err := ioutil.ReadFile(tempfile)
	if err != nil {
		log.Fatalln(err)
	}

	var output string = strings.ReplaceAll(string(input), old, new)

	err = ioutil.WriteFile(tempfile, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func copy(src string) (int64, error) {

	fmt.Println(src)

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(tmpDir + "/" + filepath.Base(src))
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func createMachineTarball(machine string) {
	// Files which to include in the tar.gz archive
	var files []string

	fs, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range fs {
		files = append(files, tmpDir+"/"+f.Name())
	}

	// Create output file
	out, err := os.Create(machine + ".tar.gz")
	if err != nil {
		log.Fatalln("Error writing archive:", err)
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = createArchive(files, out)
	if err != nil {
		log.Fatalln("Error creating archive:", err)
	}

	fmt.Println("Archive created successfully")
}

func createArchive(files []string, buf io.Writer) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

func createBucket() {
	bucket := viper.GetString("machine-s3-bucket")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("machine-s3-region"))},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		exitErrorf("Unable to create bucket %q, %v", bucket, err)
	}

	// Wait until bucket is created before finishing
	fmt.Printf("Waiting for bucket %q to be created...\n", bucket)

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

}

func uploadFileToBucket(machine string) {
	bucket := viper.GetString("machine-s3-bucket")

	var filename = machine + ".tar.gz"

	file, err := os.Open(filename)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}

	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("machine-s3-region"))},
	)
	if err != nil {
		log.Fatal(err)
	}

	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		// Print the error and exit.
		exitErrorf("Unable to upload %q to %q, %v", filename, bucket, err)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
