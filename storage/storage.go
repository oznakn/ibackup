package storage

import (
	"bytes"
	"io/ioutil"
	"log"
	"context"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client
var ctx context.Context

func Init() {
	var err error

	ctx = context.Background()

	endpoint := "odroid.oznakn.com:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Minio client created.")
}


func Upload(name string, data []byte) bool {
	reader := bytes.NewReader(data)

	_, err := minioClient.PutObject(ctx, "photos", name, reader, int64(len(data)), minio.PutObjectOptions{})

	if err != nil {
		return false
	}

	return true
}

func Get(name string) ([]byte, error) {
	result, err := minioClient.GetObject(ctx, "photos", name, minio.GetObjectOptions{})
	defer result.Close()

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(result)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetURL(name string) (*url.URL, error) {
	result, err := minioClient.PresignedGetObject(ctx, "photos", name, time.Second * 60 * 15, nil)

	if err != nil {
		return nil, err
	}

	return result, nil
}