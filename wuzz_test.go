package main

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"testing"
)

func TestDuplicateRequestHeadersAreSent(t *testing.T) {
	headers, err := parseRequestHeaders("X-Custom: one\nX-Custom: two")
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodGet, "http://example.test", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header = headers
	dump, err := httputil.DumpRequestOut(request, false)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Count(dump, []byte("X-Custom: one\r\n")) != 1 || bytes.Count(dump, []byte("X-Custom: two\r\n")) != 1 {
		t.Fatalf("request does not contain both X-Custom headers:\n%s", dump)
	}
}

func TestBuildMultipartBody(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "payload.txt")
	if err := os.WriteFile(filePath, []byte("file contents"), 0600); err != nil {
		t.Fatal(err)
	}

	body, contentType, err := buildMultipartBody("message=hello+world\nupload=@" + filePath)
	if err != nil {
		t.Fatal(err)
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("invalid multipart Content-Type %q: %v", contentType, err)
	}
	if mediaType != "multipart/form-data" {
		t.Fatalf("media type = %q, want multipart/form-data", mediaType)
	}
	boundary := params["boundary"]
	if boundary == "" {
		t.Fatal("multipart Content-Type has no boundary")
	}
	if !bytes.HasSuffix(body.Bytes(), []byte("--"+boundary+"--\r\n")) {
		t.Fatal("multipart body has no terminating boundary")
	}

	form, err := multipart.NewReader(bytes.NewReader(body.Bytes()), boundary).ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("cannot parse multipart body: %v", err)
	}
	defer form.RemoveAll()

	if values := form.Value["message"]; len(values) != 1 || values[0] != "hello world" {
		t.Fatalf("message values = %q, want [hello world]", values)
	}
	files := form.File["upload"]
	if len(files) != 1 {
		t.Fatalf("upload files = %d, want 1", len(files))
	}
	upload, err := files[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer upload.Close()
	uploaded, err := io.ReadAll(upload)
	if err != nil {
		t.Fatal(err)
	}
	if string(uploaded) != "file contents" {
		t.Fatalf("uploaded file = %q, want %q", uploaded, "file contents")
	}
}
