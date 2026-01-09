package server

import (
	minioClient "CloudStorageProject-FileServer/internal/minio"
	"CloudStorageProject-FileServer/pkg/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// TODO: поверки бакетов надо сделать в функциях всех
func getFileFunc(w http.ResponseWriter, r *http.Request) {
	// пример запроса GET /client/api/v1/get-file?api=apikey&filename=minecraft.png
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	api := r.URL.Query().Get("api")
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "filename is required", http.StatusBadRequest)
		return
	}
	Minio := r.Context().Value("minio").(*minioClient.MinioClient)

	fileMinio, err := Minio.GetOne(api, filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	stat, errStat := fileMinio.Stat()
	if errStat != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	io.Copy(w, fileMinio)
	return
}

func storeFilesFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	api := r.URL.Query().Get("api")
	minio := r.Context().Value("minio").(*minioClient.MinioClient)

	// Важно: НЕ используем ParseMultipartForm!
	// Вместо этого создаем парсер с лимитом, но без загрузки в память
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Error creating multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}

	var uploaded []string
	var errors []string

	// Читаем части multipart формы по очереди
	for {
		part, errNext := reader.NextPart()
		if errNext == io.EOF {
			break
		}
		if errNext != nil {
			errors = append(errors, fmt.Sprintf("Error reading part: %v", errNext))
			continue
		}

		// Проверяем, что это файл (а не поле формы)
		if part.FileName() == "" {
			part.Close()
			continue
		}

		// Создаем временный файл для партишиона
		tempFile, errTemp := os.CreateTemp("", "upload-*")
		if errTemp != nil {
			part.Close()
			errors = append(errors, fmt.Sprintf("Error creating temp file for %s: %v", part.FileName(), errTemp))
			continue
		}
		tempFileName := tempFile.Name()

		// Копируем данные из part во временный файл
		fileSize, errCopy := io.Copy(tempFile, part)
		part.Close()
		tempFile.Close()

		if errCopy != nil {
			os.Remove(tempFileName)
			errors = append(errors, fmt.Sprintf("Error saving %s: %v", part.FileName(), errCopy))
			continue
		}

		// Теперь открываем временный файл для чтения и загружаем в MinIO
		fileForUpload, errOpen := os.Open(tempFileName)
		if errOpen != nil {
			os.Remove(tempFileName)
			errors = append(errors, fmt.Sprintf("Error reopening %s: %v", part.FileName(), errOpen))
			continue
		}

		contentType := "application/octet-stream"

		filePartition := models.FileMinio{
			FileName:    part.FileName(),
			Reader:      fileForUpload,
			Size:        fileSize,
			ContentType: contentType,
		}

		uploadErr := minio.CreateOne(api, filePartition)

		// Закрываем и удаляем временный файл
		fileForUpload.Close()
		os.Remove(tempFileName)

		if uploadErr != nil {
			errors = append(errors, fmt.Sprintf("Error uploading %s: %v", part.FileName(), uploadErr))
			continue
		}

		uploaded = append(uploaded, part.FileName())
	}

	// Получаем список файлов
	fileList, errList := minio.FilesList(api)
	if errList != nil {
		fileList = []models.FileWebResponse{}
	}

	// Формируем ответ
	response := models.CreateFileResponse{
		Status:        200,
		Message:       fmt.Sprintf("Successfully uploaded %d files", len(uploaded)),
		NewFiles:      fileList,
		UploadedFiles: uploaded,
	}

	if len(errors) > 0 {
		response.Message = fmt.Sprintf("Uploaded %d files with %d errors", len(uploaded), len(errors))
	}

	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(response)
	w.Write(bytes)
}

func deleteFilesFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	api := r.URL.Query().Get("api")
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "filename is required", http.StatusBadRequest)
		return
	}
	minio := r.Context().Value("minio").(*minioClient.MinioClient)

	errDelete := minio.Delete(api, filename)
	if errDelete != nil {
		http.Error(w, "Error", http.StatusNotFound)
		return
	}
	fileList, errList := minio.FilesList(api)
	if errList != nil {
		http.Error(w, "Error", http.StatusNotFound)
		return
	}

	response := models.CreateFileResponse{
		Status:   200,
		Message:  "success",
		NewFiles: fileList,
	}
	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(response)
	w.Write(bytes)
	return
}

func getFilesListFunc(w http.ResponseWriter, r *http.Request) {
	// пример запроса: POST /client/api/v1/get-files-list?api=api_key
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	api := r.URL.Query().Get("api")
	minio := r.Context().Value("minio").(*minioClient.MinioClient)

	files, err := minio.FilesList(api)
	if err != nil {
		http.Error(w, "Error", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(files)
	w.Write(bytes)
	return
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	TemplatePath := r.Context().Value("tmplPath").(string)
	apikey, is := r.Cookie("apikey")
	if is != nil {
		http.ServeFile(w, r, TemplatePath+"/index.html")
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/client/api/v1/storage?api=%s", apikey.Value), http.StatusFound)
	return
}

func storagePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TemplatePath := r.Context().Value("tmplPath").(string)
	apikey := r.URL.Query().Get("api")
	if apikey == "" || apikey == "undefined" {
		http.Redirect(w, r, "/index", http.StatusFound)
		return
	}
	http.ServeFile(w, r, TemplatePath+"/storage.html")
	return
}

func zeroPath(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/index", http.StatusFound)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	health := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "OK",
		Message: "Сервер запущен и нормально функционирует",
	}
	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(health)
	w.Write(bytes)
	return
}
