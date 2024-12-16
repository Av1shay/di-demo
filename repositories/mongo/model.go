package mongo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Item struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	Value     string             `bson:"value"`
	Version   int                `bson:"version"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}
