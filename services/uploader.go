package services

import (
	"mime/multipart"
)

type S3Uploader interface {
	UploadImage(file multipart.File, fileName string) (string, error)
}
