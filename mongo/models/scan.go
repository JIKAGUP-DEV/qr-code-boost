package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Scan struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	QRCodeId primitive.ObjectID `bson:"qrCodeId"`
	Location Location           `bson:"location"`
	ScanedAt time.Time          `bson:"scanedAt"`
}
