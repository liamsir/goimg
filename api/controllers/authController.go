package controllers

import (
	"encoding/json"
	"fmt"
	"imgserver/api/models"
	u "imgserver/api/utils"
	"net/http"
	"time"

	"github.com/dchest/passwordreset"
	recaptcha "github.com/dpapathanasiou/go-recaptcha"
	"github.com/julienschmidt/httprouter"
	"github.com/tomasen/realip"
)

func init() {
	recaptcha.Init("6LfzCYgUAAAAAOIi_gUzVvhMA52WmFIgrJ-Mz3fB")
}

var CreateAccount = func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	grecaptcha := p.ByName("grecaptcha")

	if grecaptcha == "" || grecaptcha == "-1" {
		u.Respond(w, u.Message(false, "Missing CAPTCHA token."))
		return
	}
	clientIP := realip.FromRequest(r)
	result, errr := recaptcha.Confirm(clientIP, grecaptcha)
	if errr != nil || !result {
		u.Respond(w, u.Message(false, "Invalid CAPTCHA token."))
		return
	}

	account := &models.User{}
	err := json.NewDecoder(r.Body).Decode(account) //decode the request body into struct and failed if any error occur
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}
	resp := account.Create() //Create account
	u.Respond(w, resp)
}

var GetUserProfile = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := r.Context().Value("user").(uint)

	data := models.GetUser(uint(user))
	allowedDomains := models.GetDomainsFor(uint(user))
	resp := u.Message(true, "success")
	resp["data"] = data
	resp["allowedDomains"] = allowedDomains
	u.Respond(w, resp)
}

var Authenticate = func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	grecaptcha := p.ByName("grecaptcha")

	if grecaptcha == "" || grecaptcha == "-1" {
		u.Respond(w, u.Message(false, "Missing CAPTCHA token."))
		return
	}
	clientIP := realip.FromRequest(r)
	result, errr := recaptcha.Confirm(clientIP, grecaptcha)
	if errr != nil || !result {
		u.Respond(w, u.Message(false, "Invalid CAPTCHA token."))
		return
	}

	account := &models.User{}
	err := json.NewDecoder(r.Body).Decode(account) //decode the request body into struct and failed if any error occur
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	resp := models.Login(account.Email, account.Password)
	u.Respond(w, resp)
}

var AuthenticateWithRefreshToken = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	account := &models.User{}
	err := json.NewDecoder(r.Body).Decode(account) //decode the request body into struct and failed if any error occur
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	resp := models.LoginWithRefreshToken(account.Email, account.RefreshToken)
	u.Respond(w, resp)
}

var forgotPasswordSecret = "oCgv-dCHy2eGcsjeGFdR9K-uEdIqhYT8rdK1Tq2Cdqyb0m4YbWXy8XXxL1FVno6VTiYpMgAA8_bsp-Q9Yk_xww"

var ForgotPassword = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data := &models.ForgotPasswordViewModel{}
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	secret := []byte(forgotPasswordSecret)
	pwdval, err := getPasswordHash(data.Email)
	if err != nil {
		// user doesn't exists, abort
		return
	}
	// Generate reset token that expires in 12 hours
	token := passwordreset.NewToken(data.Email, 12*time.Hour, pwdval, secret)
	fmt.Println(token)
	u.Respond(w, u.Message(true, "Link with token is in its way."))
}

func getPasswordHash(login string) ([]byte, error) {
	user := models.GetUserWithEmail(login)
	if user == nil {
		return nil, fmt.Errorf("Error")
	}
	return []byte(user.Password), nil
}

var ResetPassword = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	data := &models.ResetPasswordViewModel{}
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}
	secret := []byte(forgotPasswordSecret)
	login, err := passwordreset.VerifyToken(data.Token, getPasswordHash, secret)
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	user := models.User{Email: login, Password: data.Password}
	resp := user.UpdatePassword()

	u.Respond(w, resp)
}
