package controllers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"os"
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
		log.Println(err)
		if c.AbortWithError(408, errors.New("could not bind json to shopper pointer")) != nil {
			c.Errors.JSON()
		}
		return
	}

	sameEmails, err := shopperCollection.CountDocuments(ctx, bson.D{{Key: "email", Value: shopper.Email}})
	if err != nil {
		log.Println(err)
		if c.AbortWithError(400, errors.New("could not search for duplicate documents")) != nil {
			return
		}
		return
	}

	if sameEmails > 0 {
		c.AbortWithStatusJSON(409, gin.H{"unsuccessful": "honey this email already exists"})
	}

	samePhone, err := shopperCollection.CountDocuments(ctx, bson.D{{Key: "phone", Value: shopper.Phone}})
	if err != nil {
		log.Println(err)
		if c.AbortWithError(400, errors.New("could not perform search for duplicate phone numbers")) != nil {
			return
		}
		return
	}

	if samePhone > 0 {
		log.Println("This phone already exists")
		//c.JSON(409, "phone already exists in db")
		if c.AbortWithError(409, errors.New("this phone number exists in the database")) != nil {
			return
		}
		return
	}

	securePassword := helpers.HashPassword(shopper.Password)
	shopper.Password = securePassword

	shopper.CreatedAt, err = time.Parse(time.RFC850, time.Now().Format(time.RFC850))
	if err != nil {
		log.Println("Could not set createdAt time", err)
		if c.AbortWithError(400, errors.New("could not set createdAt time")) != nil {
			return
		}
		return
	}

	shopper.ID = primitive.NewObjectID()

	shopper.UserID = shopper.ID.Hex()

	result, err := shopperCollection.InsertOne(ctx, shopper)

	if err != nil {

		if c.AbortWithError(500, errors.New("could not insert new user into collection")) != nil {
			return
		}
		log.Panic(err)
		return
	}
	c.JSON(201, gin.H{"success": result.InsertedID})

}

func LoginShopper(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var shopper models.User
	if err := c.BindJSON(&shopper); err != nil {
		if err := c.AbortWithError(400, errors.New("could not parse and set json body to shopper struct")); err != nil {
			return
		}
		log.Println("Could not bind json to shopper struct", err)
		return
	}
	var matchingUser models.User
	if err := shopperCollection.FindOne(ctx,
		bson.D{{Key: "email", Value: shopper.Email}}).Decode(&matchingUser); err != nil {
		log.Println("Could not perform search to get matching user")
	}

	emailExists, err := shopperCollection.CountDocuments(ctx, bson.D{{Key: "email", Value: shopper.Email}})
	if err != nil {
		log.Println("Unable to count documents", err)
		return
	}
	if emailExists < 1 {
		if err := c.AbortWithError(404, errors.New("could not find matching email")); err != nil {
			return
		}
		log.Println("Could not find any matching email", err)
		return
	}

	valid, msg := helpers.VerifyPassword(matchingUser.Password, shopper.Password)
	if msg != "Password is a match" {
		if err := c.AbortWithError(404, errors.New("unable to verify password")); err != nil {
			c.JSON(404, err)
		}
		log.Println("Unable to verify password", err)
		return
	}
	if !valid {
		if err := c.AbortWithError(400, errors.New("password in incorrect")); err != nil {
			log.Println(err)
		}
		log.Println("Passwords did not match")
		return
	}

	/////////////////////////////////////////////
	// Credentials are correct and can be used
	/////////////////////////////////////////////
	savedRefreshTokenCookie, err := c.Request.Cookie("RefreshToken")
	if err == http.ErrNoCookie {
		log.Println(``)
	}
	if err != nil {
		log.Println(``)
	}
	log.Println("The refresh cookie is", savedRefreshTokenCookie.Value)
	newAuthToken := helpers.GenerateNewAccessToken(c, matchingUser)
	if err := helpers.CreateCookiesForTokens(c, newAuthToken, savedRefreshTokenCookie.Value); err != nil {
		log.Println("Unable to create cookies for newly generated tokens", err)
		if err := c.AbortWithError(400, errors.New("unable to create cookies for newly generated tokens")); err != nil {
			return
		}
	}
	if err == nil {
		c.JSON(200, "done!")
		return
	}

	signedAuthToken, signedRefreshToken, err := helpers.GenerateAllTokens(matchingUser)

	if err != nil {
		if err := c.AbortWithError(400, errors.New("unable to generate auth and refresh tokens")); err != nil {
			return
		}
		log.Println("Unable to generate tokens", err)
	}
	if err := helpers.CreateCookiesForTokens(c, signedAuthToken, signedRefreshToken); err != nil {
		if err := c.AbortWithError(419, errors.New("could not create cookies for successfully created jwt tokens")); err != nil {
			return
		}
		return
	}

	matchingUser.UpdatedAt, err = time.Parse(time.RFC850, time.Now().Format(time.RFC850))

	if err != nil {
		log.Println("Error setting time of update", err)
		return
	}

	matchingId, err := primitive.ObjectIDFromHex(matchingUser.UserID)
	if err != nil {
		c.JSON(404, "unable to get matching ID from hex value")
		log.Println("Unable to extract _id from hex", err)
		return
	}

	if err := shopperCollection.FindOne(ctx, bson.D{{Key: "_id", Value: matchingId},
		{Key: "email", Value: matchingUser.Email}, {Key: "phone", Value: matchingUser.Phone}}).
		Decode(&matchingUser); err != nil {
		log.Println("Error matching user, some credentials may be incorrect", err)
		return
	}

	claims, result := ValidateToken(signedAuthToken)
	if result == "" {
		log.Println("Seems alright")
	}
	log.Println("Email", claims.UserId)
	//correctUserID, err := primitive.ObjectIDFromHex(claims.UserId)
	//if err != nil {
	//	log.Println("We need to check the user ID again")
	//}
	//if err := shopperCollection.FindOne(ctx, bson.D{{"_id", correctUserID}}).Decode(&matchingUser); err != nil {
	//	log.Println(err)
	//}
}
func ValidateToken(token string) (claims *helpers.SignedDetails, result string) {
	finalToken, err := jwt.ParseWithClaims(token, &helpers.SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET_KEY")), nil
	})
	if err != nil {
		log.Println(err)
	}
	claims, ok := finalToken.Claims.(*helpers.SignedDetails)
	if !ok {
		log.Println("Unable to parse token")
	}
	return claims, result
}

func GetUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var matchingUser models.User
	signedAuthToken, err := c.Request.Cookie("AuthToken")
	if err == http.ErrNoCookie {
		log.Println("That cookie cannot be found")
		if err := c.AbortWithError(404, errors.New("auth Cookie not found")); err != nil {
			return
		}
	}

	claims, result := ValidateToken(signedAuthToken.Value)

	if result != "" {
		log.Println("We have work to do")
	}
	log.Println("We are back in business")

	log.Println("Matching User ID", claims.UserId)

	correctUserID, err := primitive.ObjectIDFromHex(claims.UserId)

	if err != nil {
		log.Println("We need to check the user ID again")
	}
	if err := shopperCollection.FindOne(ctx, bson.D{{"_id", correctUserID}}).Decode(&matchingUser); err != nil {
		log.Println(err)
	}
	log.Println(matchingUser)
	c.JSON(200, matchingUser)
}

func LogoutShopper(c *gin.Context) {
	err := helpers.NullifyAllCookies(c)
	if err != nil {
		log.Println(err)
	}
	c.JSON(301, "Successfully created your delete cookies")
}
