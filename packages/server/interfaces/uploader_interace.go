package interfaces

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"bitbucket.org/hofng/hofApp/pkg/uploader"
)

func UploadFile(w http.ResponseWriter, r *http.Request, svc interface{}) {
	//TODO: Enhanced Logging
	fmt.Println("File Upload Endpoint Hit")

    r.ParseMultipartForm(10 << 20)

    file, handler, err := r.FormFile("image_file")
	
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
		encodeResult(w, err, http.StatusInternalServerError)
        return
    }
	defer file.Close()

	format := "image: {filename: %s size:%d header:%v}"
	out := fmt.Sprintf(format, handler.Filename, handler.Size, handler.Header)

	log.Print(out)

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {		
		encodeResult(w, err, http.StatusInternalServerError)
	}

	fileHandler := uploader.FileHandler{
		FileName: handler.Filename, 
		FileSize: handler.Size,
		File: buf.Bytes(),
	}

	result, err := svc.(uploader.Service).UploadFile(r.Context(), fileHandler)

	if err != nil {
		EncodeJSONError(r.Context(), err, w)
		return
	}
	encodeResult(w, result, http.StatusOK)
}

