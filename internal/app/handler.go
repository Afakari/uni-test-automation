package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register new user
func RegisterHandler(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if creds.Username == "" || creds.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password required"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption error"})
		return
	}

	if _, loaded := Users.LoadOrStore(creds.Username, string(hashed)); loaded {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}

// Login existing user
func LoginHandler(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// --- REQUIRED: Check for null/empty fields ---
	if creds.Username == "" || creds.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password required"})
		return
	}
	// ---------------------------------------------

	hashed, ok := Users.Load(creds.Username)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashed.(string)), []byte(creds.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": creds.Username,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// Create new Todo
func CreateTodoHandler(c *gin.Context) {
	username, _ := c.Get("username")
	var req Todo
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	newTodo := &Todo{
		ID:        GenerateID(),
		Title:     req.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	curr, ok := Todos.Load(username)
	if !ok {
		Todos.Store(username, []*Todo{newTodo})
		c.JSON(http.StatusCreated, newTodo)
		return
	}

	slice := curr.([]*Todo)
	slice = append(slice, newTodo)
	Todos.Store(username, slice)
	c.JSON(http.StatusCreated, newTodo)
}

// Get all Todos
func GetTodosHandler(c *gin.Context) {
	username, _ := c.Get("username")
	v, ok := Todos.Load(username)
	if !ok {
		c.JSON(http.StatusOK, []Todo{})
		return
	}

	ptrSlice := v.([]*Todo)
	out := make([]Todo, 0, len(ptrSlice))
	for _, p := range ptrSlice {
		if p != nil {
			out = append(out, *p)
		}
	}
	c.JSON(http.StatusOK, out)
}

// Get single Todo
func GetTodoHandler(c *gin.Context) {
	username, _ := c.Get("username")
	id := c.Param("id")

	v, ok := Todos.Load(username)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}
	ptrSlice := v.([]*Todo)

	for _, p := range ptrSlice {
		if p != nil && p.ID == id {
			c.JSON(http.StatusOK, p)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
}

// Update Todo (intentional race)
func UpdateTodoHandler(c *gin.Context) {
	username, _ := c.Get("username")
	id := c.Param("id")

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	v, ok := Todos.Load(username)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}
	ptrSlice := v.([]*Todo)

	var found *Todo
	for _, p := range ptrSlice {
		if p != nil && p.ID == id {
			found = p
			break
		}
	}

	if found == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	if req.Title != nil {
		found.Title = *req.Title // race
	}
	if req.Completed != nil {
		found.Completed = *req.Completed // race
	}

	c.JSON(http.StatusOK, found)
}

// Delete Todo
func DeleteTodoHandler(c *gin.Context) {
	username, _ := c.Get("username")
	id := c.Param("id")

	v, ok := Todos.Load(username)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}
	ptrSlice := v.([]*Todo)

	idx := -1
	for i, p := range ptrSlice {
		if p != nil && p.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	newSlice := append(ptrSlice[:idx], ptrSlice[idx+1:]...)
	Todos.Store(username, newSlice)
	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
}
