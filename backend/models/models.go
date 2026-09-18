package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

type PollOption struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Options   []PollOption       `bson:"options" json:"options"`
	AuthorID  primitive.ObjectID `bson:"author_id" json:"author_id"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// Vote is the durable record of an individual vote, persisted to MongoDB
// in addition to the live Redis counter used for real-time broadcasting.
type Vote struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID           primitive.ObjectID `bson:"poll_id" json:"poll_id"`
	OptionID         string             `bson:"option_id" json:"option_id"`
	VoterFingerprint string             `bson:"voter_fingerprint" json:"-"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
}

// PollUpdatePayload is broadcast over Redis Pub/Sub and forwarded to
// WebSocket clients whenever a vote is cast.
type PollUpdatePayload struct {
	PollID string         `json:"poll_id"`
	Counts map[string]int `json:"counts"`
}
