package server

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
)

// MaxUploadSize is the Max file size.
const MaxUploadSize = 1024 * 1024 // 1MB

// StartServer starts the web server.
func StartServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/upload", getFileUploadHandler())
	mux.HandleFunc("/file", getUploadedFileHandler())

	return http.ListenAndServe(":2113", mux)
}

var fileContent = make(map[string][]byte)

func getUploadedFileHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Invalid key", http.StatusBadRequest)
			return
		}
		content := fileContent[key]
		if len(content) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(fileContent[key])
	}
}

func getFileUploadHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		log.Println("File being uploaded")
		r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
		defer r.Body.Close()

		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Invalid key", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("data")
		if err != nil {
			log.Printf("Failed to get file: %v\n", err)
			http.Error(w, "Failed to get file", http.StatusInternalServerError)
			return
		}
		if header.Size > MaxUploadSize {
			http.Error(w, "File size is too large", http.StatusBadRequest)
			return
		}

		buf, err := ioutil.ReadAll(file)
		if err != nil {
			log.Printf("Failed to read body: %v\n", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		fileContent[key] = buf
		w.WriteHeader(http.StatusOK)
	}
}
