package main

import (
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/api/option"
	"io"
	"mime/multipart"
	"net/http"
)

var (
	credentials = "./Keys/dpacks-3e038-9865a5b29f91.json"
	bucketName  = "dpacks-3e038.appspot.com"
)

//const (
//	host     = ""
//	port     = 5432
//	user     = ""
//	password = ""
//	dbname   = ""
//)

//const (
//	SecretKey = ""
//)

//func ValidateJWT() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		// Get the JWT string from the header
//		tokenString := c.GetHeader("Authorization")
//
//		// Validate the token
//		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//			// Make sure the token method is HMAC
//			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//				return nil, jwt.ErrSignatureInvalid
//			}
//			return []byte(SecretKey), nil
//		})
//
//		if err != nil || !token.Valid {
//			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
//			c.Abort()
//			return
//		}
//
//		// If the token is valid, extract the claims
//		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
//			// You can access the user data in the claims
//			// For example, if the user data is stored in the "user" field
//			userID := claims["id"]
//			// You can then store the user data in the Gin context
//			c.Set("userid", userID)
//		}
//
//		// Proceed with the request
//		c.Next()
//	}
//}

func main() {
	r := gin.Default()

	// CORS middleware configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}

	r.Use(cors.New(config))

	// File upload endpoint
	r.POST("/write", uploadFile)
	r.DELETE("/:filename", removeFile)

	// File view endpoint
	//r.GET("/view/:filename", viewFile)

	// File delete endpoint
	//r.DELETE("/delete/:filename", ValidateJWT(), removeFile)

	r.Run(":4004")
}

// Function to connect to the database
//func dbConnect() *sql.DB {
//	// Connect to the "bank" database.
//	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
//		host, port, user, password, dbname))
//	if err != nil {
//		log.Fatal(err)
//	}
//	return db
//}

// Function to check if the file is a JSON file
func isJSON(file *multipart.FileHeader) bool {
	allowedTypes := map[string]bool{
		"application/json": true,
		// Add more allowed JSON types as needed
	}

	contentType := file.Header.Get("Content-Type")
	return allowedTypes[contentType]
}

// =====================================================================================================================
// Function to upload a file to Firebase Storage and insert a record in the database
func uploadFile(c *gin.Context) {

	// Get the user POST data from the Gin context filename as string
	rawFilename := c.PostForm("filename")

	// make filename as string
	filename := fmt.Sprintf("%v", rawFilename)

	// Check if the user data exists
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No filename"})
		return
	}

	// Limit the maximum file size to 4MB
	maxSize := int64(5 << 20) // 5MB
	err := c.Request.ParseMultipartForm(maxSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds the limit"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
		return
	}

	// Validate file type (allow only json)
	//if !isJSON(file) {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JSON files are allowed"})
	//	return
	//}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	// Initialize Firebase app
	opt := option.WithCredentialsFile(credentials)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize Firebase app"})
		return
	}

	// Create a Firebase Storage client
	client, err := app.Storage(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Firebase Storage client"})
		return
	}

	// Create a Storage bucket reference
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bucket reference"})
		return
	}

	// Create an object in the bucket with the specified content type
	obj := bucket.Object(filename)
	w := obj.NewWriter(context.Background())
	w.ContentType = "application/json"

	// Write data to the object
	if _, err := io.Copy(w, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file to Firebase Storage"})
		return
	}
	defer w.Close()

	//insertUploadData(1, filename)

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "image": filename})
}

// convertToInteger converts an interface{} to an int
func convertToInteger(value interface{}) int {
	var result int
	switch value.(type) {
	case int:
		result = value.(int)
	case int8:
		result = int(value.(int8))
	case int16:
		result = int(value.(int16))
	case int32:
		result = int(value.(int32))
	case int64:
		result = int(value.(int64))
	case uint:
		result = int(value.(uint))
	case uint8:
		result = int(value.(uint8))
	case uint16:
		result = int(value.(uint16))
	case uint32:
		result = int(value.(uint32))
	case uint64:
		result = int(value.(uint64))
	case float32:
		result = int(value.(float32))
	case float64:
		result = int(value.(float64))
	}
	return result
}

// Function to delete a file from Firebase Storage and remove the record from the database
func removeFile(c *gin.Context) {
	// Get the user data from the Gin context
	//userIdRaw, exists := c.Get("userid")
	//if !exists {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "No user data"})
	//	return
	//}

	//var user interface{} = userIdRaw
	//userId := convertToInteger(user)

	// Get the filename from the request
	filename := c.Param("filename")

	// Check if the user is authorized to delete the file
	//if !isAuthorizedToDelete(userId, filename) {
	//	c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to delete this file"})
	//	return
	//}

	// Initialize Firebase app
	opt := option.WithCredentialsFile(credentials)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create a Firebase Storage client
	client, err := app.Storage(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the bucket reference
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the object from the bucket
	obj := bucket.Object(filename)
	if err := obj.Delete(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

//
//// Function to check if the user is authorized to delete the file
//func isAuthorizedToDelete(userId int, filename string) bool {
//	db := dbConnect()
//
//	// Perform a prepared statement with a parameterized query to check authorization
//	row := db.QueryRow("SELECT user_id FROM temp_user_files WHERE file_name = $1", filename)
//	var storedUserId int
//	if err := row.Scan(&storedUserId); err != nil {
//		fmt.Println(err)
//		return false
//	}
//
//	defer func(db *sql.DB) {
//		err := db.Close()
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//	}(db)
//
//	// Check if the user ID matches the stored user ID
//	return userId == storedUserId
//}
//
//// Function to insert a record in the database
//func insertUploadData(userId int, fileName string) {
//
//	db := dbConnect()
//
//	// Perform a prepared statement with parameterized query
//	stmt, err := db.Prepare(
//		"INSERT INTO temp_user_files(user_id, file_name, date_time) VALUES ($1, $2, now())")
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	defer func(stmt *sql.Stmt) {
//		err := stmt.Close()
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//	}(stmt)
//
//	// Execute the prepared statement with parameters
//	_, err = stmt.Exec(userId, fileName)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	defer func(db *sql.DB) {
//		err := db.Close()
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//	}(db)
//}

// Function to generate a unique filename inspired by Facebook's algorithm
//func generateUniqueFileName(originalFilename string) string {
//	ext := filepath.Ext(originalFilename)
//
//	// Generate a random string (you can adjust the length as needed)
//	randomStr := generateRandomString(8)
//
//	// Append a timestamp (Unix nanoseconds)
//	timestamp := time.Now().UnixNano()
//
//	// Combine all parts to create a unique filename
//	return fmt.Sprintf("%s_%d%s", randomStr, timestamp, ext)
//}

// Function to generate a random string of the specified length
//func generateRandomString(length int) string {
//	bytes := make([]byte, length)
//	_, err := rand.Read(bytes)
//	if err != nil {
//		panic(err)
//	}
//	return hex.EncodeToString(bytes)
//}

// Function to view a file from Firebase Storage
//func viewFile(c *gin.Context) {
//	filename := c.Param("filename")
//
//	// Authenticate user and check authorization from the database
//	// Add your authentication and authorization logic here
//
//	// Initialize Firebase app
//	opt := option.WithCredentialsFile(credentials)
//	app, err := firebase.NewApp(context.Background(), nil, opt)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	// Create a Firebase Storage client
//	client, err := app.Storage(context.Background())
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	// Get the bucket reference
//	bucket, err := client.Bucket(bucketName)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	// Get the object from the bucket
//	obj := bucket.Object(filename)
//	r, err := obj.NewReader(context.Background())
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//	defer r.Close()
//
//	// Read the contents of the storage.Reader into a []byte
//	fileContent, err := io.ReadAll(r)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	// Stream the file to the client
//	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", filename))
//	c.Data(http.StatusOK, "application/octet-stream", fileContent)
//}
