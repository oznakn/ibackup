package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	rice "github.com/GeertJohan/go.rice"
	"github.com/rs/xid"
	"log"
	"lukechampine.com/blake3"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const PageSize = 10

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	offset := 0

	if val := r.FormValue("page"); val != "" {
		page, err := strconv.Atoi(val)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		offset = (page - 1) * PageSize

		if offset < 0 {
			offset = 0
		}
	}

	var images []Image
	Conn.Order("taken_at DESC").Offset(offset).Limit(PageSize).Find(&images)

	cleanStorageCache()

	result := make([]map[string]interface{}, len(images))

	for i, image := range images {
		imageURL, err := fetchImageUrl(image.Filename)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		result[i] = map[string]interface{}{
			"devicePath": image.DevicePath,
			"source": image.Source,
			"name": image.Filename,
			"takenAt": image.TakenAt,
			"uploadedAt": image.CreatedAt,
			"size": image.Size,
			"url": imageURL.String(),
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{} {
		"images": result,
	})
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	source := r.FormValue("source")
	path := r.FormValue("path")
	dateAsString := r.FormValue("date")
	if source == "" || path == "" || dateAsString == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dateAsInt, err := strconv.ParseInt(dateAsString, 10, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ext := strings.ToLower(filepath.Ext(handler.Filename))

	fileBytes, err := compressIfPossible(file, ext)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hashAsBytes := blake3.Sum512(fileBytes)
	hash := hex.EncodeToString(hashAsBytes[:])

	var foundImage *Image

	var images []Image
	Conn.Find(&images, "hash = ?", hash)

	for _, image := range images {
		storageFileBytes, err := fetchImage(image.Filename)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		if bytes.Compare(storageFileBytes, fileBytes) == 0 {
			foundImage = &image

			break
		}
	}

	if foundImage == nil {
		filename := xid.New().String() + filepath.Ext(handler.Filename)

		uploadImage(filename, fileBytes)

		date := time.Unix(dateAsInt, 0)

		image := Image{
			Source: source,
			DevicePath: path,
			Filename: filename,
			Hash: hash,
			Size: uint(len(fileBytes)),
			TakenAt: date,
		}

		Conn.Create(&image)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string {
			"status": "uploaded",
			"hash": hash,
			"name": filename,
		})
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string {
			"status": "exists",
			"hash": hash,
			"name": foundImage.Filename,
		})
	}
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := r.FormValue("name")

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var image Image
	if err := Conn.Where("filename = ?", name).Delete(&image).Error; err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err := deleteImage(name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string {
		"status": "deleted",
	})
}

func startServer(bindAddress string) {
	http.HandleFunc("/api/images", HomeHandler)
	http.HandleFunc("/api/upload", UploadHandler)
	http.HandleFunc("/api/delete", DeleteHandler)
	http.Handle("/", http.FileServer(rice.MustFindBox("./static").HTTPBox()))

	log.Printf("Server started at %s.", bindAddress)
	log.Fatal(http.ListenAndServe(bindAddress, nil))
}

