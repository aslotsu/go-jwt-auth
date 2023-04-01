package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
	"user-auth/database"
	"user-auth/helpers"
	"user-auth/models"
)

var client = database.Connect()
var shopperCollection = client.Database("shop").Collection("shoppers")

func SignUpShopper(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var shopper models.User
	if err := c.BindJSON(&shopper); err != nil {
		log.Println("Could not bind json to shopper pointer", err)
		return
	}
	sameEmails, err := shopperCollection.CountDocuments(ctx, bson.D{{"email", shopper.Email}})
	if err != nil {
		log.Println("Could not search for duplicate documents", err)
		return
	}

	if sameEmails > 0 {
		log.Println("This email already exists")
		return
	}
	samePhone, err := shopperCollection.CountDocuments(ctx, bson.D{{"phone", shopper.Phone}})
	if err != nil {
		log.Println("Could not search for duplicate documents", err)
		return
	}

	if samePhone > 0 {
		log.Println("This phone already exists")
		return
	}

	securePassword := helpers.HashPassword(shopper.Password)
	shopper.Password = securePassword

	shopper.CreatedAt, err = time.Parse(time.RFC850, time.Now().Format(time.RFC850))
	if err != nil {
		log.Println("Could not set createdAt time", err)
		return
	}

	shopper.ID = primitive.NewObjectID()

	shopper.UserID = shopper.ID.Hex()

	authToken, refreshToken, err := helpers.GenerateAllTokens(shopper)
	if err != nil {
		log.Println("Could not generate tokens", err)
	}

	result, err := shopperCollection.InsertOne(ctx, shopper)

	if err != nil {
		log.Println("Could not insert into document", err)
		return
	}

	if err := helpers.CreateCookiesForTokens(c, authToken, refreshToken); err != nil {
		log.Panic("Could not create cookies", err)
		return
	}
	c.JSON(201, gin.H{"success": result.InsertedID})

}

func LoginShopper(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var shopper models.User
	if err := c.BindJSON(&shopper); err != nil {
		log.Println("Could not bind json to shopper struct", err)
		return
	}
	var matchingUser models.User
	if err := shopperCollection.FindOne(ctx,
		bson.D{{"email", shopper.Email}}).Decode(&matchingUser); err != nil {
		log.Println("Could not traverse the document", err)
		return
	}

	emailExists, err := shopperCollection.CountDocuments(ctx, bson.D{{"email", shopper.Email}})
	if err != nil {
		log.Println("Unable to count documents", err)
		return
	}
	if emailExists < 1 {
		log.Println("Could not find any matching email", err)
		return
	}

	valid, msg := helpers.VerifyPassword(matchingUser.Password, shopper.Password)
	if msg != "Password is a match" {
		log.Println("Unable to verify password", err)
		return
	}
	if !valid {
		log.Println("Passwords did not match")
		return
	}

	signedAuthToken, signedRefreshToken, err := helpers.GenerateAllTokens(matchingUser)
	if err != nil {
		log.Println("Unable to generate tokens", err)
	}
	matchingUser.UpdatedAt, err = time.Parse(time.RFC850, time.Now().Format(time.RFC850))

	if err != nil {
		log.Println("Error setting time of update", err)
		return
	}
	matchingId, err := primitive.ObjectIDFromHex(matchingUser.UserID)
	if err != nil {
		log.Println("Unable to extract _id from hex", err)
		return
	}
	helpers.UpdateAllTokens(c, signedAuthToken, signedRefreshToken)
	if err := shopperCollection.FindOne(ctx, bson.D{{"_id", matchingId}, {"email", matchingUser.Email}, {"phone", matchingUser.Phone}}).Decode(&matchingUser); err != nil {
		log.Println("Error matching user, some credentials may be incorrect", err)
		return
	}
	c.JSON(200, matchingUser)
}

func LogoutShopper(c *gin.Context) {
	helpers.NullifyAllCookies(c)
}
