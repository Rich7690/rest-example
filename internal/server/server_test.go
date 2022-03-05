package server

import (
	"bytes"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"crypto/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileUpload(t *testing.T) {
	handler := getFileUploadHandler()

	f, err := ioutil.TempFile("", "test-*")
	defer os.Remove(f.Name())

	assert.NoError(t, err)

	randomBytes := make([]byte, 64)
	_, err = rand.Read(randomBytes)
	assert.NoError(t, err)

	err = ioutil.WriteFile(f.Name(), randomBytes, fs.ModeAppend)
	assert.NoError(t, err)

	r, err := Upload("example.com?key=foo", map[string]io.Reader{"data": f})
	rw := httptest.NewRecorder()

	handler(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)

	rw = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/file?key=foo", nil)
	r.URL.Query().Set("key", "foo")

	getUploadedFileHandler()(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, randomBytes, rw.Body.Bytes())
	log.Printf("bits: %v\n", randomBytes)
}

func Upload(url string, values map[string]io.Reader) (r *http.Request, err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return nil, err
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return nil, err
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return nil, err
		}
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	return req, err
}
