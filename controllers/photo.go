package controllers

import (
	"cloud.google.com/go/storage"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/appengine"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	storageClient *storage.Client
)

// GetAllPhotos is a controller that returns a list of photos
func GetAllPhotos(c *gin.Context) {
	// Load env vars
	_ = godotenv.Load()
	bucket := os.Getenv("BUCKET_NAME")

	var err error

	// Set up GCP Storage Client
	ctx := appengine.NewContext(c.Request)

	storageClient, err = storage.NewClient(ctx, option.WithCredentialsFile("keys.json"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"error":   true,
		})
		return
	}

	// Get all photos in bucket
	query := &storage.Query{Prefix: ""}
	it := storageClient.Bucket(bucket).Objects(ctx, query)

	var photoNames []string
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		photoNames = append(photoNames, attrs.Name)
	}

	c.JSON(http.StatusOK, gin.H{
		"numberOfPhotos": len(photoNames),
		"photos":         photoNames,
	})
}

// GetPhotoByName is a controller that returns a photo by name
func GetPhotoByName(c *gin.Context) {
	// Get role from url
	photoName := c.Param("photoName")

	fmt.Println(photoName)

	_ = godotenv.Load()
	bucket := os.Getenv("BUCKET_NAME")

	var err error

	// Set up GCP Storage Client
	ctx := appengine.NewContext(c.Request)

	storageClient, err = storage.NewClient(ctx, option.WithCredentialsFile("keys.json"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"error":   true,
		})
		return
	}

	// Get photo in bucket
	photo := storageClient.Bucket(bucket).Object(photoName)

	w := photo.NewWriter(ctx)
	fmt.Println(w.ObjectAttrs.ContentType)

	// Read it back.
	r, err := photo.NewReader(ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Error Reading File",
			"error":   true,
		})
		return
	}
	defer r.Close()

	// Read file into buffer
	data, err := ioutil.ReadAll(r)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Error Reading File Data",
			"error":   true,
		})
		return
	}

	// Return photo
	c.Data(http.StatusOK, w.ContentType, data)
}

// UploadPhoto is a POST route that uploads a photo
func UploadPhoto(c *gin.Context) {
	_ = godotenv.Load()
	bucket := os.Getenv("BUCKET_NAME")

	var err error

	// Get uploaded file
	f, uploadedFile, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"error":   true,
		})
		return
	}

	// Close file at end of func
	defer f.Close()

	// Ensure Content-Type is image
	contentType := uploadedFile.Header["Content-Type"][0]
	contentTypeArr := strings.Split(contentType, "/")
	if contentTypeArr[0] != "image" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid Content-Type",
			"error":   true,
		})
		return
	}

	// Set up GCP Storage Client
	ctx := appengine.NewContext(c.Request)

	storageClient, err = storage.NewClient(ctx, option.WithCredentialsFile("keys.json"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"error":   true,
		})
		return
	}

	id := uuid.New()

	sw := storageClient.Bucket(bucket).Object(uploadedFile.Filename + id.String()).NewWriter(ctx)

	if _, err := io.Copy(sw, f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"error":   true,
		})
		return
	}

	if err := sw.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"error":   true,
		})
		return
	}

	u, err := url.Parse("/" + bucket + "/" + sw.Attrs().Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"Error":   true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "file uploaded successfully",
		"pathname": u.EscapedPath(),
	})
}

// DeletePhoto is a DELETE route that deletes a photo by name
func DeletePhoto(c *gin.Context) {
	// Get role from url
	photoName := c.Param("photoName")

	fmt.Println(photoName)

	_ = godotenv.Load()
	bucket := os.Getenv("BUCKET_NAME")

	var err error

	// Set up GCP Storage Client
	ctx := appengine.NewContext(c.Request)

	storageClient, err = storage.NewClient(ctx, option.WithCredentialsFile("keys.json"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"error":   true,
		})
		return
	}

	photo := storageClient.Bucket(bucket).Object(photoName)
	err = photo.Delete(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
			"error":   true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "file deleted successfully",
	})
}
