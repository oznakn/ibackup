package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	rice "github.com/GeertJohan/go.rice"
	"github.com/oznakn/ibackup-server/db"
	"github.com/oznakn/ibackup-server/storage"
	"github.com/rs/xid"
	"io/ioutil"
	"log"
	"lukechampine.com/blake3"
	"net/http"
	"path/filepath"
	"strconv"
)

const PageSize = 10

func Home(w http.ResponseWriter, r *http.Request) {
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

	var images []db.Image
	db.Conn.Order("created_at DESC").Offset(offset).Limit(PageSize).Find(&images)

	storage.CleanCache()

	result := make([]map[string]interface{}, len(images))

	for i, image := range images {
		imageURL, err := storage.GetURL(image.Filename)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		result[i] = map[string]interface{}{
			"devicePath": image.DevicePath,
			"source": image.Source,
			"name": image.Filename,
			"date": image.CreatedAt,
			"url": imageURL.String(),
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{} {
		"images": result,
	})
}

func Upload(w http.ResponseWriter, r *http.Request) {
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
	if source == "" || path == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fileBytes, err := ioutil.ReadAll(file)
	hashAsBytes := blake3.Sum512(fileBytes)
	hash := hex.EncodeToString(hashAsBytes[:])

	var foundImage *db.Image

	var images []db.Image
	db.Conn.Find(&images, "hash = ?", hash)

	for _, image := range images {
		storageFileBytes, err := storage.Get(image.Filename)

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

		storage.Upload(filename, fileBytes)

		image := db.Image{
			Source: source,
			DevicePath: path,
			Filename: filename,
			Hash: hash,
		}

		db.Conn.Create(&image)

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

func main() {
	db.Init()
	storage.Init()

	http.HandleFunc("/api/images", Home)
	http.HandleFunc("/api/upload", Upload)
	http.Handle("/", http.FileServer(rice.MustFindBox("./static").HTTPBox()))

	log.Println("Server started.")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
