package models

import "io"

type FileMinio struct {
	FileName    string
	Data        []byte
	Reader      io.Reader // для больших файлов, для стриминга
	ContentType string    // для стриминга
	Size        int64
}

type FileWebResponse struct {
	FileName    string `json:"file_name"`
	FileType    string `json:"file_type"`
	LastModTime string `json:"create_date"`
	FileSize    string `json:"file_size"`
}

type CreateFileResponse struct {
	Status        int               `json:"status"`
	Message       string            `json:"message"`
	NewFiles      []FileWebResponse `json:"new_files"`
	UploadedFiles []string          `json:"uploaded_files"`
}
