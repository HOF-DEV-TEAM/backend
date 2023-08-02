package uploader

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
)

func UploadFile(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(uploadFile, svc)
}

func parseFile(r *http.Request) (*FileHandler, error) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Print("Error Retrieving the File")

		return nil, err
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	format := "image: {filename: %s size:%d header:%v}"
	out := fmt.Sprintf(format, handler.Filename, handler.Size, handler.Header)

	log.Print(out)

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, err
	}

	return &FileHandler{
		FileName: handler.Filename,
		FileSize: handler.Size,
		File:     buf.Bytes(),
	}, nil
}

func uploadFile(w http.ResponseWriter, r *http.Request, svc interface{}) {
	bucketKey := r.FormValue("hof")
	if bucketKey != "" {
		bucketKey += "/"
	}

	//TODO: Enhanced Logging
	fileHandler, err := parseFile(r)
	if err != nil {
		log.Println("parsefile: ", err)
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	result, err := svc.(Service).UploadFile(r.Context(), fileHandler, bucketKey)
	if err != nil {
		log.Println("Upload file", err)
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}
