package controllers

import (
	"encoding/json"
	"imgserver/api/models"
	u "imgserver/api/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

var GetDomainsFor = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := r.Context().Value("user").(uint)
	data := models.GetDomainsFor(uint(user))
	resp := u.Message(true, "success")
	resp["data"] = data
	u.Respond(w, resp)
}

var GetDomain = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := r.Context().Value("user").(uint)
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		//The passed path parameter is not an integer
		u.Respond(w, u.Message(false, "There was an error in your request"))
		return
	}
	data := models.GetDomain(uint(user), id)
	resp := u.Message(true, "success")
	resp["data"] = data
	u.Respond(w, resp)
}
var CreateDomain = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user := r.Context().Value("user").(uint) //Grab the id of the user that send the request
	domain := &models.Domain{}

	err := json.NewDecoder(r.Body).Decode(domain)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}
	domain.UserId = user
	resp := domain.Create()
	u.Respond(w, resp)
}

var UpdateDomain = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := r.Context().Value("user").(uint) //Grab the id of the user that send the request
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		u.Respond(w, u.Message(false, "There was an error in your request"))
		return
	}
	domain := &models.Domain{}
	domain.ID = uint(id)
	err = json.NewDecoder(r.Body).Decode(domain)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}
	domain.UserId = user
	resp := domain.Update()
	u.Respond(w, resp)
}

var PatchDomain = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := r.Context().Value("user").(uint) //Grab the id of the user that send the request
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		u.Respond(w, u.Message(false, "There was an error in your request"))
		return
	}
	domain := &models.Domain{}
	domain.ID = uint(id)
	err = json.NewDecoder(r.Body).Decode(domain)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}
	domain.UserId = user
	resp := domain.Patch()
	u.Respond(w, resp)
}

var DeleteDomain = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	user := r.Context().Value("user").(uint)
	id := ps.ByName("id")

	data, err := models.DeleteDomain(int(user), id)
	if err != nil {
		resp := u.Message(false, err.Error())
		u.Respond(w, resp)
		return
	}
	resp := u.Message(true, "success")
	resp["data"] = data
	u.Respond(w, resp)
}
