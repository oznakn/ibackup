package main

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

func storageInit(endpoint string, accessKey string, secretKey string) {
	var err error

	ctx = context.Background()

	c = cache.New(time.Hour, 15*time.Minute)

	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Minio client created.")
}

func cleanStorageCache() {
	c.DeleteExpired()
}

func uploadImage(name string, data []byte) bool {
	reader := bytes.NewReader(data)

	_, err := minioClient.PutObject(ctx, "photos", name, reader, int64(len(data)), minio.PutObjectOptions{})

	if err != nil {
		return false
	}

	return true
}

func fetchImage(name string) ([]byte, error) {
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

func fetchImageUrl(name string) (*url.URL, error) {
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

func deleteImage(name string) error {
	err := minioClient.RemoveObject(ctx, "photos", name, minio.RemoveObjectOptions{})

	if err != nil {
		return err
	}

	return nil
}