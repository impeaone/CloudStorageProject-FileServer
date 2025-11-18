package server

import (
	"CloudStorageProject-FileServer/pkg/validation"
	"fmt"
	"io"
	"net/http"
	"os"
)

func getFilesFunc(w http.ResponseWriter, r *http.Request) {
	// пример запроса GET /client/api/v1/get-file?filename=minecraft.png&expires=2314321&signature=1avr4vc
	// hmac используем
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expires := r.URL.Query().Get("expires")
	signature := r.URL.Query().Get("signature")
	// Проверка, действительна ли ссылка
	if expires == "" || signature == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// TODO: нужна валидация названия файла, чтобы лишний раз os.Open не писать

	filename := r.URL.Query().Get("filename")
	validURL, errSig := validation.ValidateURL(filename, expires, signature)
	if errSig != nil || !validURL {
		http.Error(w, fmt.Sprintf("Bad Request:%v", errSig), http.StatusBadRequest)
		return
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", FilesDirectory, filename))
	if err != nil {
		//logger
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	_, errCopy := io.Copy(w, file)
	if errCopy != nil {
		//logger
		http.Error(w, errCopy.Error(), http.StatusInternalServerError)
		return
	}
	//logger info

}

// TODO: додэлать
func storeFilesFunc(w http.ResponseWriter, r *http.Request) {
	// пример запроса: POST /client/api/v1/upload-files?expires=2314321&signature=1avr4vc
	// hmac нужен
	// все файлы в теле запроса
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	api := r.FormValue("apikey")
	// Проверка подлинности ссылки по api
	expires := r.URL.Query().Get("expires")
	signature := r.URL.Query().Get("signature")
	// Проверка, действительна ли ссылка
	if expires == "" || signature == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	validURL, errSig := validation.ValidateURL(api, expires, signature)
	if errSig != nil || !validURL {
		http.Error(w, fmt.Sprintf("Bad Request:%v", errSig), http.StatusBadRequest)
		return
	}

	//TODO: поход в бд, сохранение файла,
	//TODO: создание папки если ее нет
	apiDIR := api
	// получаем файлы из формы
	files := r.MultipartForm.File["files"]

	for _, fileHend := range files {
		file, err := fileHend.Open()
		if err != nil {
			//logger
			continue
		}
		dst, errC := os.Create(fmt.Sprintf("%s/%s/%s", FilesDirectory, apiDIR, fileHend.Filename))
		if errC != nil {
			continue
		}

		io.Copy(dst, file)

		//logger
		file.Close()
		dst.Close()
	}

}

func deleteFilesFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getFilesListFunc(w http.ResponseWriter, r *http.Request) {
	// пример запроса GET /client/api/v1/?api=api_key
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

}
