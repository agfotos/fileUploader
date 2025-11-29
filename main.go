package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	http.HandleFunc("/upload", fileUploadHandler)

	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func fileUploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(250 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	dir := rand.IntN(1000)
	dirName := "uploads/" + strconv.Itoa(dir)
	multipartFormData := r.MultipartForm
	for _, fh := range multipartFormData.File["uploadfile"] {
		file, err := fh.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		defer file.Close()

		fmt.Fprintf(w, "Uploaded File: %+v\n", fh.Filename)
		fmt.Printf("File size: %d\n", fh.Size)
		fmt.Printf("MIME header: %v\n", fh.Header)

		dst, err := createFile(dirName, fh.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer dst.Close()
		/**
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "Invalid File", http.StatusBadRequest)
			return
		}
		*/

		/**
		if !isValidFileType(fileBytes) {
			http.Error(w, "Invalid File Type", http.StatusUnsupportedMediaType)
			return
		}
		*/
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		/**
		if err := uploadToS3(fileBytes, fh.Filename, "D"+strconv.Itoa(dir)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		*/
		fmt.Fprintf(w, "File uploaded to s3")
	}
	err := createZipFile(dirName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createFile(directory string, filename string) (*os.File, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}

	dst, err := os.Create(filepath.Join(directory, filename))
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func isValidFileType(file []byte) bool {
	fileTYpe := http.DetectContentType(file)
	return strings.HasPrefix(fileTYpe, "image/")
}

func uploadToS3(file []byte, filename string, directory string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}

	s3Client := s3.New(sess)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("henco.downloads"),
		Key:    aws.String(directory + "/" + filename),
		Body:   bytes.NewReader(file),
		ACL:    aws.String("public-read"),
	})
	return err
}

func createZipFile(directory string) error {
	fmt.Println("creating zip archive...")
	archive, err := os.Create("archive.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	entries, err := os.ReadDir(directory)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	fmt.Printf("Contents of directory '%s':\n", directory)
	for _, entry := range entries {
		fmt.Println("opening first file...")
		f1, err := os.Open(filepath.Join(directory, entry.Name()))
		if err != nil {
			panic(err)
		}
		defer f1.Close()

		fmt.Println("writing first file to archive...")
		w1, err := zipWriter.Create(entry.Name())
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(w1, f1); err != nil {
			panic(err)
		}
	}
	fmt.Println("closing zip archive...")
	zipWriter.Close()
	return nil
}
