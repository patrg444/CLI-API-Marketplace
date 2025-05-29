package auth

import (
    "github.com/gin-gonic/gin"
)

type User struct {
    UserID   string
    Email    string
    Username string
}

func GetUserFromContext(c *gin.Context) (*User, bool) {
    userID, exists := c.Get("user_id")
    if !exists {
        return nil, false
    }
    
    email, _ := c.Get("user_email")
    username, _ := c.Get("username")
    
    return &User{
        UserID:   userID.(string),
        Email:    email.(string),
        Username: username.(string),
    }, true
}

func IsAdmin(user *User) bool {
    // Mock implementation
    return user.Email == "admin@example.com"
}
