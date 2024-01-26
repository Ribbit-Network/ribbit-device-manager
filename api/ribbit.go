package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
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
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	store     = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	cookie    *http.Cookie
	loginUser = "username"
	loginPass = "password"
)

// session manager
func getSession(c *gin.Context) *sessions.Session {
	session, err := store.Get(c.Request, "session-name")
	if err != nil {
		fmt.Printf("Error : problem starting session -> %v\n", err.Error())
		return nil
	}

	return session
}

func Signup(c *gin.Context) {
	creds := &Credentials{
		Email:    c.Param("email"),
		Password: c.Param("password"),
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	hashedPassword := string(hashedPasswordBytes)

	user := db.User{
		Email:    creds.Email,
		Password: hashedPassword,
	}

	if err := db.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signup successful"})
}

func Signin(c *gin.Context) {
	creds := &Credentials{
		Email:    c.Param("email"),
		Password: c.Param("password"),
	}

	// TODO: set cookie
	// set cookie
	// cookie, err := c.Cookie("logged-in")

	// no cookie

	// Get the existing entry present in the database for the given username
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

	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

func Login(res http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("logged-in")

	// no cookie
	if err == http.ErrNoCookie {
		cookie = &http.Cookie{
			Name:  "logged-in",
			Value: "0",
		}
	}

	// check log in: password entered = "secret"?
	// TODO: Lookup actual user creds from storage
	if req.Method == "POST" {
		password := req.FormValue("password")
		if password == "secret" {
			cookie = &http.Cookie{
				Name:  "logged-in",
				Value: "1",
			}
		}
	}

	// if logout, then logout and destroy cookie
	if req.URL.Path == "/logout" {
		cookie = &http.Cookie{
			Name:   "logged-in",
			Value:  "0",
			MaxAge: -1,
		}
	}

	http.SetCookie(res, cookie)

	// create string with html for response
	var html string

	// not logged in
	if cookie.Value == strconv.Itoa(0) {
		html = `<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title></title>
		</head>
		<body>
		<form method="post" action="/">
			<h3>User name</h3>
			<input type="text" name="userName">
			<h3>Password</h3>
			<input type="text" name="password">
			<br>
			<input type="submit">
			<input type="submit" name="logout" value="logout">
		</form>
		</body>
		</html>`
	}

	// logged in
	if cookie.Value == strconv.Itoa(1) {
		html = `<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title></title>
		</head>
		<body>
		,h1><a href="/logout">LOGOUT</a></h1>
		</body>
		</html>`
	}

	io.WriteString(res, html) // send data to client side
}

func GetAllUsers(c *gin.Context) {
	users, err := db.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func DeleteUser(c *gin.Context) {
	email := c.Param("email")

	if err := db.DeleteUserByEmail(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func dashboardHandler(c *gin.Context) {
	session := getSession(c)

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", nil)
}

// createNewDevice adds a device to the active user account
func createNewDevice(c *gin.Context) {
	// TODO: retrieve active email

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

	r.GET("/cookie", func(c *gin.Context) {

		cookie, err := c.Cookie("gin_cookie")

		if err != nil {
			cookie = "NotSet"
			c.SetCookie("gin_cookie", "test", 3600, "/", "localhost", false, true)
		}

		fmt.Printf("Cookie value: %s \n", cookie)
	})

	// Front end handlers
	r.LoadHTMLGlob("../templates/*")
	r.Static("/static", "./static")

	// Ribbit API handlers

	http.HandleFunc("/login", Login)
	r.GET("/dashboard", dashboardHandler)
	r.POST("/signin/:email/:password", Signin)
	r.POST("/signup/:email/:password", Signup)

	// Golioth API handlers
	r.POST("/createNewDevice", createNewDevice)
	r.DELETE("/users/:email", DeleteUser)
	r.POST("/createDeviceGolioth", createDeviceNoDB) // Used exclusively for testing

	return r
}

// TODO: on app startup, run a check against the golioth API to get all devices and compare against the database
