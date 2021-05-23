// Package backup copies docker-machine certs and configurations to S3
package backup

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

var backupCmd = CreateBackupCmd()

var tmpDir = ""

// TestMode var determines if certain flows actually complete or not for unit testing
var TestMode = false

// MachineS3Bucket defines which bucket mach interacts with for storing config tarballs, pulled from `machine-s3-bucket` in .mach.conf.yaml
var MachineS3Bucket string = "mach-docker-machine-certificates"

// MachineS3Region defines which region the bucket is in, pulled from `machine-s3-region` in .mach.conf.yaml
var MachineS3Region string = "us-east-1"

// CreateBucketFirst will trigger the creation of a bucket before a backup, triggered with cli flag `-c` or `--create`
var CreateBucketFirst bool = false

// KeepTarball will trigger a clean-up of the tarball, set to true to prevent, or `-k` or `--keep-tarball`
var KeepTarball bool = false

func CreateBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup <docker-machine>",
		Short: "Takes a working docker-machine entry and stores it to an S3 bucket",
		Long: `This allows you to store the docker-machine certs bundle in an S3 bucket, 
paired with restore command, this will let you transfer docker-machines to and from
systems using the AWS API. Will require programmatic credentials with permissions to download from S`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackup(cmd, args)
		},
	}
	return cmd
}

func init() {

	TestMode = strings.HasSuffix(os.Args[0], ".test")

	viper.SetDefault("machine-s3-bucket", MachineS3Bucket)
	MachineS3Bucket = viper.GetString("machine-s3-bucket")

	viper.SetDefault("machine-s3-region", MachineS3Region)
	MachineS3Region = viper.GetString("machine-s3-region")

	backupCmd.Flags().BoolP("create", "c", CreateBucketFirst, "create the bucket before attempting backup")
	CreateBucketFirst, _ = backupCmd.Flags().GetBool("create")

	backupCmd.Flags().BoolP("keep-tarball", "k", KeepTarball, "keeps the tarball in working directory after upload")
	KeepTarball, _ := backupCmd.Flags().GetBool("keep-tarball")
	if KeepTarball {
		fmt.Println("--keep-tarball set")
	}
}

// runBackup is the main command flow, it will attempt to create an S3 bucket if
// the flag is set, then it will create a temp directory, populate it with the
// machine config, tarball it, and push to S3, and delete the temp files created
func runBackup(cmd *cobra.Command, args []string) error {

	if CreateBucketFirst {
		createBucket()
	}

	if len(args) == 1 {
		createTempDirectory()
		populateTempDir(args[0])
		createMachineTarball(args[0])
		uploadFileToBucket(args[0])

		defer os.RemoveAll(tmpDir)

		fmt.Println(args[0] + " backup complete to " + MachineS3Bucket + " bucket")

		if !KeepTarball {
			removeMachineArchive(args[0])
		}
	}

	return nil
}

// populateTempDir copies the relevant files from ~/.docker/machine/ to the tmp directory and
// triggers replaceIntTempFile
func populateTempDir(machine string) {

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	var machinedir = homedir + "/.docker/machine/machines/" + machine + "/"
	var certsdir = homedir + "/.docker/machine/certs/"

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

// replaceInTempFile performs an old -> new swap of a string against a file
// used to replace paths with template placeholders
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

// copy copies a file into the temp directory for processing
func copy(src string) (int64, error) {

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
}

// createArchive makes a tarball out of the contents of temp directory
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

// addToArchive tarballs a file
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

// createBucket will create the bucket referenced in the machine-s3-bucket string. Call this with a flag.
func createBucket() {
	bucket := MachineS3Bucket

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

// uploadFileToBucket takes the machine tarball and puts it in S3
func uploadFileToBucket(machine string) {
	bucket := MachineS3Bucket

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
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func createTempDirectory() string {
	dir, err := ioutil.TempDir("/tmp", "machine")
	if err != nil {
		log.Fatal(err)
	}

	tmpDir = dir
	return tmpDir
}

func removeMachineArchive(machine string) {
	if !KeepTarball {
		e := os.Remove(machine + ".tar.gz")
		if e != nil {
			log.Fatal(e)
		}
	}

}
