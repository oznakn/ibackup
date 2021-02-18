package main

import (
	"flag"
)

func main() {
	bindAddress := flag.String("bind", "127.0.0.1:8080", "Bind address")
	dbPath := flag.String("db", "ibackup.db", "Sqlite DB Path")
	connectAddress := flag.String("connect", "127.0.0.1:9000", "Storage Server")
	accessKey := flag.String("accessKey", "minioadmin", "Storage Server Access Key")
	secretKey := flag.String("secretKey", "minioadmin", "Storage Server Secret Key")

	flag.Parse()

	if !flag.Parsed() {
		flag.PrintDefaults()

		return
	}

	dbInit(*dbPath)
	storageInit(*connectAddress, *accessKey, *secretKey)

	startServer(*bindAddress)
}
