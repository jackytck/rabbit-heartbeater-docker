package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// UploadS3 uploads a file to s3.
func UploadS3(filePath string) {
	key, secret, region, bucket := LoadAwsCreds()
	token := ""
	creds := credentials.NewStaticCredentials(key, secret, token)
	_, err := creds.Get()
	if err != nil {
		fmt.Printf("bad credentials: %s", err)
	}
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("err opening file: %s", err)
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size) // read file content to buffer

	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	path := "/" + file.Name()
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(path),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
		CacheControl:  aws.String("max-age=0"),
	}
	_, err = svc.PutObject(params)
	if err != nil {
		fmt.Printf("bad response: %s", err)
	}
	// fmt.Printf("response %s\n", awsutil.StringValue(resp))
	LogGreen("Uploaded to S3")
}
