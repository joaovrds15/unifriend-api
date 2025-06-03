package mocks

import (
	"mime/multipart"
)

type MockS3Uploader struct {
	UploadImageFunc func(file multipart.File, fileName string) (string, error)
	DeleteImageFunc func(fileName string) error
}

func (m *MockS3Uploader) UploadImage(file multipart.File, fileName string) (string, error) {
	return m.UploadImageFunc(file, fileName)
}

func (m *MockS3Uploader) DeleteImage(fileName string) error {
	return m.DeleteImageFunc(fileName)
}
