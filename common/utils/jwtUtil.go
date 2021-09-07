package utils

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type User struct {
	ID uint64       `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Phone string `json:"phone"`
}

var user = User{
	ID:            18,
	Username: "username",
	Password: "password",
	Phone: "49123454322", //this is a random number
}

const secretKey="jdnfksdmfksd"

func Login(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
		return
	}
	//compare the user from the request, with the one we defined:
	if user.Username != u.Username || user.Password != u.Password {
		c.JSON(http.StatusUnauthorized, "Please provide valid login details")
		return
	}
	token, err := CreateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	c.JSON(http.StatusOK, token)
}

func CreateToken(userId uint64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["user_id"] = userId
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()

	tokenString, err := token.SignedString([]byte(secretKey))

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	c, e := ExtractToken(tokenString)
	fmt.Println(c,e)
	exp := GetTokenExpiration(tokenString)
	fmt.Println(exp)

	return tokenString, err
}

func ExtractToken(tokenStr string) (jwt.MapClaims, bool) {
	hmacSecret := []byte(secretKey)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// check token signing method etc
		return hmacSecret, nil
	})

	if err != nil {
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	} else {
		return nil, false
	}
}

func GetTokenExpiration(tokenString string) (int64){
	claims,ok := ExtractToken(tokenString)
	if !ok{
		return -1
	}

	//fmt.Println(claims)
	expirationFloat, ok := claims["exp"].(float64)
	if !ok{
		return -1
	}

	expirationInt := int64(expirationFloat)
	//fmt.Println(expirationInt,ok)

	return expirationInt
}

