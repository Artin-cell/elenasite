package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminIDFromContext возвращает id администратора, которого положил RequireAdmin.
func AdminIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(adminKey)
	if !ok {
		return uuid.Nil, false
	}
	s, ok := v.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
