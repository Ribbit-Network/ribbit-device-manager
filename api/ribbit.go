package api

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Initialize the session store
var store = cookie.NewStore([]byte("secret"))

func init() {
	store.Options(sessions.Options{
		MaxAge:   int(15 * 60), // 15 minutes in seconds
		Path:     "/",
		HttpOnly: true,
	})
}

func sessionExpiryMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	lastActivity, found := session.Get("lastActivity").(time.Time)

	if !found || time.Since(lastActivity) > 15*time.Minute {
		session.Clear()
		session.Save()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
		c.Abort()
		return
	}

	// Update the last activity time
	session.Set("lastActivity", time.Now())
	session.Save()
	c.Next()
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

	// After user creation
	session := sessions.Default(c)
	session.Set("email", creds.Email)
	session.Set("lastActivity", time.Now())
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "Signup successful"})
}

func Signin(c *gin.Context) {
	creds := &Credentials{
		Email:    c.Param("email"),
		Password: c.Param("password"),
	}

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

	// After successful login
	session := sessions.Default(c)
	session.Set("email", creds.Email)
	session.Set("lastActivity", time.Now())
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
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

// createNewDevice adds a device to the active user account
func createNewDevice(c *gin.Context) {
	// Access the session
	session := sessions.Default(c)

	// Retrieve the active user's email from the session
	email, ok := session.Get("email").(string)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "user email not found in session"})
		return
	}

	// Fetch the user details using the email from the session
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
		UserID:     user.ID,
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
	// TODO: resolve session manager bug
	// r.Use(sessions.Sessions("mysession", store))
	// r.Use(sessionExpiryMiddleware)
	r.POST("/signin/:email/:password", Signin)
	r.POST("/signup/:email/:password", Signup)
	r.POST("/createNewDevice", createNewDevice)
	r.DELETE("/users/:email", DeleteUser)
	r.POST("/createDeviceGolioth", createDeviceNoDB) // Used exclusively for testing
	return r
}

// TODO: on app startup, run a check against the golioth API to get all devices and compare against the database
