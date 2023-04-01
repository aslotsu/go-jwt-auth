package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id" `
	FullName  string             `bson:"fullName" json:"fullName" validate:"required,min=6,max=100"`
	Email     string             `bson:"email" json:"email" validate:"email,required"`
	Phone     string             `bson:"phone" json:"phone" validate:"required"`
	Password  string             `bson:"password" json:"password" validate:"required,min=6,max=100"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	UserID    string             `bson:"userID" json:"userID"`
	Role      string             `bson:"role" json:"role" validate:"required,eq=BUSINESS|eq=SHOPPER"`
}
