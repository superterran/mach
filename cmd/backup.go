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
	"fmt"
	"os"
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
		uploadFileToBucket(args[0])
	}

	return nil
}

func createBucket() {
	bucket := viper.GetString("machine-s3-bucket")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("machine-s3-region"))},
	)

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

func uploadFileToBucket(filename string) {
	bucket := viper.GetString("machine-s3-bucket")

	file, err := os.Open(filename)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}

	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("machine-s3-region"))},
	)

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
