package models

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	u "imgserver/api/utils"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var token_password = "Bz72tJc0s21JQlntgY2DeTi5pipCFiox"

/*
JWT claims struct
*/
type Token struct {
	UserId   uint
	Username string
	jwt.StandardClaims
}

//a struct to rep user account
type User struct {
	gorm.Model
	Email        string `json:"email"`
	Username     string `json:"username"`
	LastName     string `json:"last_name"`
	Password     string `json:"password"`
	Token        string `json:"token" sql:"-"`
	RefreshToken string `json:"refresh_token"`
	SecretKey    string `json:"secret_key"`
}

type ForgotPasswordViewModel struct {
	Email string
}

type ResetPasswordViewModel struct {
	Token    string
	Password string
}

//Validate incoming user details...
func (account *User) Validate() (map[string]interface{}, bool) {

	if !strings.Contains(account.Email, "@") {
		return u.Message(false, "Email address is required"), false
	}

	if len(account.Password) < 6 {
		return u.Message(false, "Password is required"), false
	}

	//Email must be unique
	temp := &User{}

	//check for errors and duplicate emails
	err := GetDB().Table("users").Where("email = ? || username = ?", account.Email, account.Username).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error. Please retry"), false
	}
	if temp.Email != "" {
		return u.Message(false, "Email address already in use by another user."), false
	}

	return u.Message(false, "Requirement passed"), true
}

func (account *User) UpdatePassword() map[string]interface{} {

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)

	GetDB().Model(&account).Where("email = ?", account.Email).Update("password", account.Password)

	response := u.Message(true, "Password has been updated")
	return response
}

func (account *User) Create() map[string]interface{} {

	if resp, ok := account.Validate(); !ok {
		return resp
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	secretKey, err := generateSecreyKey()
	if err != nil {
		return u.Message(false, err.Error())
	}
	account.SecretKey = secretKey
	refreshToken, err := generateSecreyKey()
	if err != nil {
		return u.Message(false, err.Error())
	}
	account.RefreshToken = refreshToken
	GetDB().Create(account)

	if account.ID <= 0 {
		return u.Message(false, "Failed to create account, connection error.")
	}

	//os.Getenv("token_password")

	//Create new JWT token for the newly registered account
	tk := &Token{UserId: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(token_password))
	account.Token = tokenString
	account.Password = "" //delete password

	response := u.Message(true, "Account has been created")
	response["account"] = account
	return response
}

func generateSecreyKey() (string, error) {
	key := make([]byte, 64)

	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("Error generating secret key.")
	}
	secretKey := base64.RawURLEncoding.EncodeToString(key)
	return secretKey, nil
}

func Login(email, password string) map[string]interface{} {

	account := &User{}
	err := GetDB().Table("users").Where("email = ?", email).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.Message(false, "Email address not found")
		}
		return u.Message(false, "Connection error. Please retry")
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
		return u.Message(false, "Invalid login credentials. Please try again")
	}
	//Worked! Logged In
	account.Password = ""

	//Create JWT token
	tk := &Token{UserId: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(token_password))
	account.Token = tokenString //Store the token in the response

	resp := u.Message(true, "Logged In")
	resp["account"] = account
	return resp
}

func LoginWithRefreshToken(email, refreshToken string) map[string]interface{} {

	account := &User{}
	err := GetDB().Table("users").Where("(refresh_token = '') is not true AND email = ? AND refresh_token = ?", email, refreshToken).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.Message(false, "User not found")
		}
		return u.Message(false, "Connection error. Please retry")
	}

	newRefreshToken, err := generateSecreyKey()
	if err != nil {
		return u.Message(false, err.Error())
	}
	account.RefreshToken = newRefreshToken
	db.Model(&account).Update("refresh_token", newRefreshToken)

	//Worked! Logged In
	account.Password = ""

	//Create JWT token
	tk := &Token{UserId: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(token_password))
	account.Token = tokenString //Store the token in the response

	resp := u.Message(true, "Logged In")
	resp["account"] = account
	return resp
}

func GetUser(u uint) *User {

	acc := &User{}
	GetDB().Table("users").Where("id = ?", u).First(acc)
	if acc.Email == "" { //User not found!
		return nil
	}

	acc.Password = ""
	return acc
}

func GetUserWith(username string, secretkey string) *User {

	acc := &User{}
	GetDB().Table("users").Where("username = ? AND secret_key = ?", username, secretkey).First(acc)
	if acc.Email == "" { //User not found!
		return nil
	}

	acc.Password = ""
	return acc
}

func GetUserWithUsername(username string) *User {

	acc := &User{}
	GetDB().Table("users").Where("username = ?", username).First(acc)
	if acc.Email == "" { //User not found!
		return nil
	}

	acc.Password = ""
	return acc
}

func GetUserWithEmail(email string) *User {
	acc := &User{}
	GetDB().Table("users").Where("email = ?", email).First(acc)
	if acc.Email == "" { //User not found!
		return nil
	}
	return acc
}
