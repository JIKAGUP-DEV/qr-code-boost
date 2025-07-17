package models

type Location struct {
	Type        string    `bson:"type"`        // Será sempre "Point"
	Coordinates []float64 `bson:"coordinates"` // Array com [longitude, latitude]
}
