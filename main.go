package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// CustomTransport to handle path rewrite
type CustomTransport struct {
	Transport http.RoundTripper
	BaseURL   string
}

func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Path = t.BaseURL + req.URL.Path
	return t.Transport.RoundTrip(req)
}

type ConnectionString struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BaseURL         string
	BucketName      string
	ObjectName      string
	conn            *minio.Client
}

// NewConnectionString 建構函數，初始化並建立與 Minio 的連線
func NewConnectionString(endpoint, accessKeyID, secretAccessKey, baseURL, bucketName, objectName string) *ConnectionString {
	conn := &ConnectionString{
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		BaseURL:         baseURL,
		BucketName:      bucketName,
		ObjectName:      objectName,
	}

	if conn.connectMinio() != nil {
		return nil
	} else {
		return conn
	}

}

func (cs *ConnectionString) connectMinio() error {
	// Custom HTTP transport settings to ignore certificate verification
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// Initialize custom transport to handle base URL
	transport := &CustomTransport{
		Transport: customTransport,
		BaseURL:   cs.BaseURL,
	}

	useSSL := true
	// Initialize MinIO client object with custom transport
	minioClient, err := minio.New(cs.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cs.AccessKeyID, cs.SecretAccessKey, ""),
		Secure:    useSSL,
		Transport: transport,
	})
	if err != nil {
		log.Fatalln(err)
		return err
	} else {
		cs.conn = minioClient
		return nil
	}

}
func (cs *ConnectionString) MakeBucket() error {

	err := cs.conn.MakeBucket(context.Background(), cs.BucketName, minio.MakeBucketOptions{Region: "us-east-2"})
	if err != nil {
		exists, errBucketExists := cs.conn.BucketExists(context.Background(), cs.BucketName)
		if errBucketExists == nil && exists {
			fmt.Printf("Bucket %s already exists\n", cs.BucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		fmt.Printf("Successfully created bucket %s\n", cs.BucketName)
	}

	return nil
}

func (cs *ConnectionString) Putfile() error {

	content := "Hello, World! ltcloud upload test by edward!!!"
	_, err := cs.conn.PutObject(context.Background(), cs.BucketName, cs.ObjectName, strings.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: "text/plain"})
	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Printf("Successfully uploaded %s to %s/%s\n", cs.ObjectName, cs.BucketName, cs.ObjectName)
	}

	return nil
}

func (cs *ConnectionString) CreatePutURL() error {

	// Generate presigned URL for the object
	expiry := time.Duration(7*24) * time.Hour
	presignedURL, err := cs.conn.PresignedPutObject(context.Background(), cs.BucketName, cs.ObjectName, expiry)
	if err != nil {
		log.Fatalln(err)
	}

	// Manually adjust the URL to add the /myminio prefix
	presignedURLStr := presignedURL.String()
	customPresignedURL := strings.Replace(presignedURLStr, cs.Endpoint, cs.Endpoint+cs.BaseURL, 1)

	fmt.Printf("Presigned PUT URL:\n %s\n", customPresignedURL)

	return err
}
func (cs *ConnectionString) CreateGetURL() error {

	// Generate presigned URL for the object
	expiry := time.Duration(7*24) * time.Hour
	presignedURL, err := cs.conn.PresignedGetObject(context.Background(), cs.BucketName, cs.ObjectName, expiry, nil)
	if err != nil {
		log.Fatalln(err)
	}

	// Manually adjust the URL to add the /myminio prefix
	presignedURLStr := presignedURL.String()
	customPresignedURL := strings.Replace(presignedURLStr, cs.Endpoint, cs.Endpoint+"/storage", 1)

	fmt.Printf("Presigned GET URL:\n %s\n", customPresignedURL)

	return err

}

func main() {

	test := NewConnectionString(
		"sat-k8s-pix.baby.juiker.net", // 測試空的 Endpoint 錯誤處理
		"admin",
		"REborn9527",
		"/storage/",
		"test2",
		"test4.txt",
	)

	if test != nil {
		test.MakeBucket()
	} else {
		return
	}

	fmt.Print(test)

	if test != nil {
		test.Putfile()
	}

	if test != nil {
		test.CreateGetURL()
	}

	if test != nil {
		test.CreatePutURL()
	}

}
