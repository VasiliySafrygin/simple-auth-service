package function

import (
	"auth-service/models"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

type Claims struct {
	UserId string `json:"user_id"`
	UserName string `json:"username"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	jwt.StandardClaims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func ComparePassword(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}
	return true
}

func CreateJwtToken(user models.User) (string, error) {
	var err error
	atClaims := Claims{
		UserId: user.Id,
		UserName: user.UserName,
		FirstName: user.FirstName,
		LastName: user.LastName,
		MiddleName: user.MiddleName,
		StandardClaims: jwt.StandardClaims{
			Audience:  []string{},
			ExpiresAt: jwt.At(time.Now().Add(time.Hour * 24)),
			ID:        user.Id,
			IssuedAt:  jwt.At(time.Now()),
			Issuer:    "authService",
			NotBefore: &jwt.Time{},
			Subject:   "test",
		},
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return "", err
	}
	return token, nil
}

//func RenewJwtToken(token string) (string, error) {
//	userId, err := VerifyToken(token)
//	if err != nil {
//		log.Panic("Unable verify token", err)
//		return "", err
//	}
//	return CreateJwtToken(userId)
//}

func VerifyToken(tokenString string) (id string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		log.Println("Unable verify token", err)
		//log.Fatal(err)
		return "", err
	}

	claims, _ := token.Claims.(*Claims)

	if token.Valid {
		return claims.UserId, nil

	} else {
		return "", errors.New("Not authorized")
	}
}
