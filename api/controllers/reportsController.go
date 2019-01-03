package controllers

import (
	"fmt"
	"imgserver/api/models"
	u "imgserver/api/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

var GetReportFor = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	user := r.Context().Value("user").(uint)

	startStr := ps.ByName("start")
	endStr := ps.ByName("end")

	start, end, err := parseStartEndDate(startStr, endStr)
	if err != nil {
		u.Respond(w, u.Message(false, err.Error()))
		return
	}

	summary := models.GetReportFor(uint(user), start, end)
	resp := u.Message(true, "success")
	resp["summary"] = summary
	u.Respond(w, resp)
}

var GetLogsFor = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	user := r.Context().Value("user").(uint)
	fmt.Println(user)
	// fileId, err := strconv.Atoi(ps.ByName("fileId"))
	page, err := strconv.Atoi(ps.ByName("page"))
	if err != nil {
		//The passed path parameter is not an integer
		u.Respond(w, u.Message(false, "There was an error in your request"))
		return
	}

	startStr := ps.ByName("start")
	endStr := ps.ByName("end")

	start, end, err := parseStartEndDate(startStr, endStr)
	if err != nil {
		u.Respond(w, u.Message(false, err.Error()))
		return
	}

	logs := models.GetLogs(uint(user), uint(page), start, end)
	resp := u.Message(true, "success")
	resp["logs"] = logs
	u.Respond(w, resp)
}

var parseStartEndDate = func(startStr string, endStr string) (time.Time, time.Time, error) {
	layout := "2006-01-02"
	var start = time.Now().AddDate(0, -1, 0)
	var end = time.Now()

	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			return start, end, err
		}
		start = t
	}

	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			return start, end, err
		}
		end = t
	}
	return start, end, nil
}
