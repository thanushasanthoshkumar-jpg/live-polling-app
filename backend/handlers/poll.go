package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"live-polling-backend/db"
	"live-polling-backend/models"
)

type createPollRequest struct {
	Title   string   `json:"title" binding:"required,min=3,max=200"`
	Options []string `json:"options" binding:"required,min=2,max=10,dive,required,min=1,max=120"`
}

func randomOptionID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// CreatePoll validates and persists a new poll. Protected by AuthRequired.
func CreatePoll(c *gin.Context) {
	var req createPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extra validation: trim, reject empty/duplicate option text.
	seen := map[string]bool{}
	options := make([]models.PollOption, 0, len(req.Options))
	for _, raw := range req.Options {
		text := strings.TrimSpace(raw)
		if text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "poll options cannot be empty"})
			return
		}
		key := strings.ToLower(text)
		if seen[key] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "poll options must be unique"})
			return
		}
		seen[key] = true
		options = append(options, models.PollOption{ID: randomOptionID(), Text: text})
	}

	userIDHex := c.GetString("userID")
	authorID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	poll := models.Poll{
		Title:     strings.TrimSpace(req.Title),
		Options:   options,
		AuthorID:  authorID,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.PollsCollection().InsertOne(ctx, poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create poll"})
		return
	}
	poll.ID = res.InsertedID.(primitive.ObjectID)

	// Initialize the Redis hash so vote counts start at 0 for every option.
	pollIDHex := poll.ID.Hex()
	fields := make(map[string]interface{}, len(options))
	for _, opt := range options {
		fields[opt.ID] = 0
	}
	if len(fields) > 0 {
		if err := db.RedisClient.HSet(ctx, db.PollVotesKey(pollIDHex), fields).Err(); err != nil {
			// Non-fatal: counts will lazily default to 0 on first read/vote.
		}
	}

	c.JSON(http.StatusCreated, poll)
}

// ListMyPolls returns polls created by the authenticated user.
func ListMyPolls(c *gin.Context) {
	userIDHex := c.GetString("userID")
	authorID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := db.PollsCollection().Find(ctx, bson.M{"author_id": authorID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch polls"})
		return
	}
	defer cursor.Close(ctx)

	polls := []models.Poll{}
	if err := cursor.All(ctx, &polls); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode polls"})
		return
	}

	c.JSON(http.StatusOK, polls)
}

// GetPoll returns poll metadata plus live vote counts. Public endpoint --
// this is what powers the shared /poll/:id voting page.
func GetPoll(c *gin.Context) {
	idHex := c.Param("id")
	pollID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var poll models.Poll
	if err := db.PollsCollection().FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	counts, err := getLiveCounts(ctx, idHex, poll.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch vote counts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"poll":   poll,
		"counts": counts,
	})
}

// getLiveCounts reads the current vote counts from Redis, defaulting any
// missing option to 0 (e.g. if the hash hasn't been initialized yet).
func getLiveCounts(ctx context.Context, pollIDHex string, options []models.PollOption) (map[string]int, error) {
	raw, err := db.RedisClient.HGetAll(ctx, db.PollVotesKey(pollIDHex)).Result()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, len(options))
	for _, opt := range options {
		counts[opt.ID] = 0
	}
	for k, v := range raw {
		n, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		counts[k] = n
	}
	return counts, nil
}
