package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate this module. Any variable of this type will have access
// to all the methods with the reciever *Tools
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

// RandomString returns a string of random characters og length n. using randomString
// as the source for the string
func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

// UploadedFile is a struct used to save information about an uploaded file
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

// UploadedFiles handles the process of uploading files via HTTP request
func (t *Tools) UploadedFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	// Check if upload directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("upload directory %s does not exist", uploadDir)
	}

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}

	err := r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("the uploaded file is too big")
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFile, err := t.processFile(hdr, uploadDir, renameFile)
			if err != nil {
				return uploadedFiles, err
			}
			uploadedFiles = append(uploadedFiles, uploadedFile)
		}
	}

	return uploadedFiles, nil
}

// processFile handles the processing of a single uploaded file
func (t *Tools) processFile(hdr *multipart.FileHeader, uploadDir string, renameFile bool) (*UploadedFile, error) {
	var uploadedFile UploadedFile
	infile, err := hdr.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening uploaded file: %w", err)
	}
	defer infile.Close()

	buff := make([]byte, 512)
	_, err = infile.Read(buff)
	if err != nil {
		return nil, fmt.Errorf("error reading file header: %w", err)
	}

	// Check file type permissions
	allowed := false
	fileType := http.DetectContentType(buff)

	if len(t.AllowedFileTypes) > 0 {
		for _, x := range t.AllowedFileTypes {
			if strings.EqualFold(fileType, x) {
				allowed = true
				break
			}
		}
	} else {
		allowed = true
	}

	if !allowed {
		return nil, errors.New("the uploaded filetype is not permitted")
	}

	// Reset file pointer to beginning
	_, err = infile.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("error resetting file pointer: %w", err)
	}

	uploadedFile.OriginalFileName = hdr.Filename

	if renameFile {
		uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
	} else {
		uploadedFile.NewFileName = hdr.Filename
	}

	outfile, err := os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName))
	if err != nil {
		return nil, fmt.Errorf("error creating destination file: %w", err)
	}
	defer outfile.Close()

	fileSize, err := io.Copy(outfile, infile)
	if err != nil {
		return nil, fmt.Errorf("error copying file contents: %w", err)
	}

	uploadedFile.FileSize = fileSize

	return &uploadedFile, nil
}
