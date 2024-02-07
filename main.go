package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load env variables
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup gin app
	r := gin.Default()
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.MaxMultipartMemory = 8 << 20

	// Setup s3 upload
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("erro ao carregar configuração AWS: %v", err)
		return
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/", func(c *gin.Context) {
		// Get the file
		file, err := c.FormFile("image")

		if err != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": "Failed to upload image",
			})
			return
		}
		// Save the file
		f, openErr := file.Open()
		if openErr != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": "Failed to upload image",
			})
			return
		}
		result, uploadErr := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String("go-with-upload"),
			Key:    aws.String(file.Filename),
			Body:   f,
			ACL:    "public-read",
		})
		fmt.Println(result)
		if uploadErr != nil {
			log.Printf("erro ao fazer upload para o AWS S3: %v", uploadErr)
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": "Falha ao fazer upload na AWS",
			})
			return
		}

		// err = c.SaveUploadedFile(file, "assets/uploads/"+file.Filename)
		// if err != nil {
		// 	c.HTML(http.StatusOK, "index.html", gin.H{
		// 		"error": "Failed to upload image",
		// 	})
		// 	return
		// }
		// Render the page
		// c.HTML(http.StatusOK, "index.html", gin.H{
		// 	"image": "/assets/uploads/" + file.Filename,
		// })
		c.HTML(http.StatusOK, "index.html", gin.H{
			"image": result.Location,
		})
	})

	r.Run()
}
