package storage

import (
	"bytes"
	"context"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client
var ctx context.Context
var c *cache.Cache

func Init() {
	var err error

	ctx = context.Background()

	c = cache.New(time.Hour, 15*time.Minute)

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

func CleanCache() {
	c.DeleteExpired()
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
	cached, found := c.Get(name)

	if found {
		return cached.(*url.URL), nil
	}

	result, err := minioClient.PresignedGetObject(ctx, "photos", name, time.Second * 60 * 60, nil)

	if err != nil {
		return nil, err
	}

	c.Set(name, result, cache.DefaultExpiration)

	return result, nil
}