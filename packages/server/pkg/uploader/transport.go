package uploader

import (
	"bytes"	
	"fmt"
	"io"
	"log"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
)

func UploadFile(w http.ResponseWriter, r *http.Request, svc interface{}) {
	//TODO: Enhanced Logging
	fmt.Println("File Upload Endpoint Hit")

    r.ParseMultipartForm(10 << 20)

    file, handler, err := r.FormFile("image_file")
	
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
		http_helper.EncodeJSONError(r.Context(), err, w)
        return
    }
	defer file.Close()

	format := "image: {filename: %s size:%d header:%v}"
	out := fmt.Sprintf(format, handler.Filename, handler.Size, handler.Header)

	log.Print(out)

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {		
		http_helper.EncodeJSONError(r.Context(), err, w)
	}

	fileHandler := FileHandler{
		FileName: handler.Filename, 
		FileSize: handler.Size,
		File: buf.Bytes(),
	}

	result, err := svc.(Service).UploadFile(r.Context(), fileHandler)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)	
}

