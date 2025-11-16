package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
)

type ActionHandler struct {
	action interfaces.Action
	redis  interfaces.Cache
}

func (h *ActionHandler) Reaction(c *gin.Context) {
	userValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userValue.(uuid.UUID)
	videoId, _ := uuid.Parse(c.Param("id"))

	reaction := c.Query("reaction")
	if reaction != "like" && reaction != "dislike" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reaction"})
		return
	}
	redisKey := fmt.Sprintf("user_action:%s:%s", user, videoId)
	current := h.redis.Get(redisKey)
	if current == reaction {
		if err := h.action.ClearReaction(user, videoId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.redis.Set(redisKey, "none", 6*time.Hour)
		h.redis.Decr(fmt.Sprintf("%s_count:%s", reaction, videoId))
		c.JSON(http.StatusOK, gin.H{"message": "reaction removed"})
		return
	}
	if err := h.action.Update(user, videoId, reaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.redis.Set(redisKey, reaction, 6*time.Hour)
	h.redis.Incr(fmt.Sprintf("%s_count:%s", reaction, videoId))

	if current == "like" && reaction == "dislike" {
		h.redis.Decr(fmt.Sprintf("like_count:%s", videoId))
	} else if current == "dislike" && reaction == "like" {
		h.redis.Decr(fmt.Sprintf("dislike_count:%s", videoId))
	}
	c.JSON(http.StatusOK, gin.H{"message": "reaction updated"})

}
