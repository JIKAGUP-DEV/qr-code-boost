package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QRCode struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Slug      string             `bson:"slug"`
	Link      string             `bson:"link"`
	Location  Location           `bson:"location"`
	UserId    string             `bson:"userId"`
	CreatedAt time.Time          `bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty"`
}
