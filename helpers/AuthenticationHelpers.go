package helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"time"
	"user-auth/models"
)

type SignedDetails struct {
	FullName string
	Email    string
	Phone    string
	UserId   string
	UserType string
	jwt.RegisteredClaims
}

func HashPassword(rawPassword string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), 14)
	if err != nil {
		log.Panic(err)
		return "axe"
	}
	if string(hashedPassword) == "" {
		log.Println("Hashed password is blank")
	}
	return string(hashedPassword)
}

func GenerateAllTokens(shopper models.User) (string, string, error) {
	authClaims := &SignedDetails{
		FullName: shopper.FullName,
		Email:    shopper.Email,
		Phone:    shopper.Phone,
		UserId:   shopper.UserID,
		UserType: shopper.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(time.Minute * 30),
			},
		},
	}
	refreshClaims := &SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(time.Hour * 168),
			},
		},
	}

	signedAuthToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims).SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		recover()
		log.Panic("Could not generate auth token", err)
		return "", "", err
	}
	signedRefreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		recover()
		log.Panic("Could not generate refresh token", err)
		return "", "", err
	}
	return signedAuthToken, signedRefreshToken, err
}

//func UpdateAllTokens(c *gin.Context, authToken, refreshToken string) {
//	authCookie := http.Cookie{
//		Name:  "AuthToken",
//		Value: authToken,
//		//Domain:   "https://trains-git-main-aslotsu.vercel.app/",
//		MaxAge:   300,
//		SameSite: http.SameSiteNoneMode,
//		Path:     "/",
//		HttpOnly: false,
//		Secure:   false,
//	}
//
//	refreshCookie := http.Cookie{
//		Name:   "RefreshToken",
//		Value:  refreshToken,
//		MaxAge: 3600,
//		//Domain:   "https://trains-git-main-aslotsu.vercel.app/",
//		SameSite: http.SameSiteNoneMode,
//		Path:     "/",
//		HttpOnly: false,
//		Secure:   false,
//	}
//
//	http.SetCookie(c.Writer, &authCookie)
//	http.SetCookie(c.Writer, &refreshCookie)
//}

func CreateCookiesForTokens(c *gin.Context, authToken, refreshToken string) error {
	authCookie := http.Cookie{
		Name:   "AuthToken",
		Value:  authToken,
		Path:   "/",
		MaxAge: 300,
		//Domain:   "https://trains-git-main-aslotsu.vercel.app/",
		SameSite: http.SameSiteNoneMode,
		HttpOnly: true,
		Secure:   true,
	}
	refreshCookie := http.Cookie{
		Name:   "RefreshToken",
		Value:  refreshToken,
		Path:   "/",
		MaxAge: 3600,
		//Domain:   "https://trains-git-main-aslotsu.vercel.app/",
		SameSite: http.SameSiteNoneMode,
		HttpOnly: true,
		Secure:   true,
	}

	http.SetCookie(c.Writer, &authCookie)
	http.SetCookie(c.Writer, &refreshCookie)

	return nil
}

func NullifyAllCookies(c *gin.Context) {

	authCookie := http.Cookie{
		Name:    "AuthToken",
		Value:   "",
		Expires: time.Now().Add(-1000 * time.Hour),
	}
	refreshCookie := http.Cookie{
		Name:    "RefreshToken",
		Value:   "",
		Expires: time.Now().Add(-1000 * time.Hour),
	}

	http.SetCookie(c.Writer, &authCookie)
	http.SetCookie(c.Writer, &refreshCookie)
}

func VerifyPassword(hashed, password string) (bool, string) {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		recover()
		log.Panic(err)

		return false, "Password is not a match"
	}
	return true, "Password is a match"
}
