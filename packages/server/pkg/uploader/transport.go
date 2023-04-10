package uploader

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
)

func UploadFile(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(uploadFile, svc)
}

func parseFile(r *http.Request) (*FileHandler, error) {
	r.ParseMultipartForm(10 << 20)

    file, handler, err := r.FormFile("resource_file")
	
    if err != nil {
        log.Print("Error Retrieving the File")
		
        return nil, err
    }
	defer file.Close()

	format := "image: {filename: %s size:%d header:%v}"
	out := fmt.Sprintf(format, handler.Filename, handler.Size, handler.Header)

	log.Print(out)

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {		
		return nil, err
	}

	return  &FileHandler{
		FileName: handler.Filename, 
		FileSize: handler.Size,
		File: buf.Bytes(),
	}, nil 
}

func uploadFile(w http.ResponseWriter, r *http.Request, svc interface{}) {
	bucketKey := r.FormValue("bucket_key")

	if bucketKey != "" {
		bucketKey += "/"
	}

	//TODO: Enhanced Logging
	fileHandler, err := parseFile(r)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	result, err := svc.(Service).UploadFile(r.Context(), fileHandler, bucketKey)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)	
}

