package api

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rosricard/ribbitDeviceManager/db"
	"golang.org/x/crypto/bcrypt"
)

var (
	projectID = os.Getenv("GOLIOTH_PROJECT_ID")
	baseURL   = "https://api.golioth.io"
	apiKey    = os.Getenv("GOLIOTH_API_KEY")
)

type Device struct {
	ID           string
	Name         string
	PreSharedKey string
	UserID       string
	ProjectID    string
	CreatedAt    time.Time
}

type Credentials struct {
	Email    string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func signup(c *gin.Context) {
	var creds *Credentials

	// Bind the form data to the user struct
	if err := c.Bind(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	hashedPassword := string(hashedPasswordBytes)

	userdb := db.User{
		Email:    creds.Email,
		Password: hashedPassword,
	}

	if err := db.CreateUser(userdb); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.Request.Method == "POST" {
		creds.Email = c.PostForm("email")
		creds.Password = c.PostForm("password")

		// Assuming validation is successful, render a new page to the user
		// TODO: move the html to a file under templates
		html := `<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>Sign Up Success</title>
		</head>
		<body>
			<h2>Sign Up Success!</h2>
			<p>Thank you for signing up, ` + creds.Email + `.</p>
			<p><a href="/">Return to Login</a></p>
		</body>
		</html>`

		io.WriteString(c.Writer, html)
		return
	}

	// If not a POST request or form not submitted yet, render the sign-up form
	c.HTML(http.StatusOK, "signup.html", nil)

	// Render a success message or redirect to another page
	c.HTML(http.StatusOK, "signup.html", gin.H{"message": "User signed up successfully"})
}

// login to the app
func login(c *gin.Context) {
	// use cookies to track if user is logged in. This is required by the app for user device association
	cookie, err := c.Request.Cookie("logged-in")

	// no cookie
	if err == http.ErrNoCookie {
		cookie = &http.Cookie{
			Name:  "logged-in",
			Value: "0",
		}
	}

	// check log in: password entered matches what's in the db?
	if c.Request.Method == "POST" {

		// get user inputs from front end
		creds := &Credentials{
			Email:    c.PostForm("email"),
			Password: c.PostForm("password"),
		}

		// Get the existing entry present in the database for the given email
		user, err := db.GetUserByEmail(creds.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		storedCreds := &Credentials{
			Password: user.Password,
			Email:    user.Email,
		}

		if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// set cookie to 1 if user creds match and show that the user is logged in
		if creds.Password == storedCreds.Password && creds.Email == storedCreds.Email {
			cookie = &http.Cookie{
				Name:  "logged-in",
				Value: "1",
			}
			//TODO: direct to page showing that user successfully logged in
		}
	}

	// Once logged in, redirect to the home page if user logs out
	if c.Request.URL.Path == "/logout" {
		cookie = &http.Cookie{
			Name:   "logged-in",
			Value:  "0",
			MaxAge: -1,
		}
		// TODO: direct to page showing that user successfully logged out
	}

	http.SetCookie(c.Writer, cookie)

	// not logged in
	if cookie.Value == strconv.Itoa(0) {
		c.HTML(http.StatusOK, "login.html", nil)
		//TODO: serve error to the user about why login failed
		return
	}

	// logged in
	if cookie.Value == strconv.Itoa(1) {
		c.HTML(http.StatusOK, "loggedin.html", nil)
		return
	}

}

// getAllUsers from db
func getAllUsers(c *gin.Context) {
	users, err := db.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// deleteUser from db
func deleteUser(c *gin.Context) {
	email := c.Param("email")

	if err := db.DeleteUserByEmail(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// TODO: what happens to the devices associated with the user?
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// createNewDevice adds a device to the active user account
func createNewDevice(c *gin.Context) {
	// TODO: retrieve active email from cookie store

	// Fetch the user details using the email from the session
	email := "username"
	user, err := db.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//create device
	d, err := createGoliothDevice()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//create private key
	psk, err := createPSK(d.DeviceId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	device := db.DeviceDB{
		DeviceID:   d.DeviceId,
		DeviceName: d.Name,
		DevicePSK:  psk.PreSharedKey,
		UserEmail:  user.Email,
		ProjectID:  d.ProjectID,
		CreatedAt:  psk.CreatedAt,
	}

	err1 := db.CreateDevice(device)
	if err1 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deviceID": d.DeviceId, "psk": psk.PreSharedKey, "email": email})
}

// createDevice creates a new device in golioth and returns the device id and psk. Does not save to ribbit db
func createDeviceNoDB(c *gin.Context) {
	// create device
	device, err := createGoliothDevice()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	// create private key for device
	pskData, err := createPSK(device.DeviceId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"deviceID": device.DeviceId, "psk": pskData.Identity})

}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Front end handlers
	r.LoadHTMLGlob("../templates/*")
	r.Static("/static", "./static")

	// Ribbit API handlers
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.POST("/login", login)
	r.GET("/logout", login)
	r.POST("/signup", signup)
	r.DELETE("/users/:email", deleteUser)

	// Golioth API handlers
	r.POST("/createNewDevice", createNewDevice)
	r.POST("/createDeviceGolioth", createDeviceNoDB) // Used exclusively for testing

	return r
}

// TODO: on app startup, run a check against the golioth API to get all devices within a given project
// and compare against the database to ensure that the database is up to date with the golioth data
