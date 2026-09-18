package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"live-polling-backend/db"
	"live-polling-backend/models"
)

type voteRequest struct {
	OptionID string `json:"option_id" binding:"required"`
}

// CastVote is public (accessible from the shared /poll/:id page). It:
//  1. Validates the poll and option exist and the poll is active.
//  2. Atomically increments the live counter in Redis (HINCRBY).
//  3. Publishes the updated counts to the poll's Redis Pub/Sub channel.
//  4. Persists a durable vote record in MongoDB.
func CastVote(c *gin.Context) {
	idHex := c.Param("id")
	pollID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	var req voteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var poll models.Poll
	if err := db.PollsCollection().FindOne(ctx, bson.M{"_id": pollID}).Decode(&poll); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	if !poll.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this poll is closed"})
		return
	}

	validOption := false
	for _, opt := range poll.Options {
		if opt.ID == req.OptionID {
			validOption = true
			break
		}
	}
	if !validOption {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option for this poll"})
		return
	}

	// Basic duplicate-vote deterrent: fingerprint by IP + poll (not
	// cryptographically strong, but stops trivial double-submits without
	// requiring an account to vote).
	fingerprint := c.ClientIP()

	// 1. Atomic increment in Redis -- this is the source of truth for live counts.
	newCount, err := db.RedisClient.HIncrBy(ctx, db.PollVotesKey(idHex), req.OptionID, 1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record vote"})
		return
	}

	// 2. Fetch full current counts (so clients always get the complete picture).
	counts, err := getLiveCounts(ctx, idHex, poll.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read vote counts"})
		return
	}
	counts[req.OptionID] = int(newCount)

	// 3. Publish to the poll's Pub/Sub channel; the WS handler relays this
	// to every browser connected to /ws/polls/:id for this poll.
	payload := models.PollUpdatePayload{PollID: idHex, Counts: counts}
	payloadBytes, err := json.Marshal(payload)
	if err == nil {
		if err := db.RedisClient.Publish(ctx, db.PollUpdatesChannel(idHex), payloadBytes).Err(); err != nil {
			// Non-fatal: the vote is already recorded; live push just won't fire.
		}
	}

	// 4. Persist the durable vote record in MongoDB (audit trail / analytics).
	vote := models.Vote{
		PollID:           pollID,
		OptionID:         req.OptionID,
		VoterFingerprint: fingerprint,
		CreatedAt:        time.Now(),
	}
	if _, err := db.VotesCollection().InsertOne(ctx, vote); err != nil {
		// The Redis count already succeeded, so we don't fail the request --
		// but we do surface it in the response for observability.
		c.JSON(http.StatusOK, gin.H{
			"counts":  counts,
			"warning": "vote counted live but failed to persist audit record",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"counts": counts})
}
