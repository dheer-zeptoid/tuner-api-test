package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"net/http"
)

type timeInRequiredFormat struct {
	Hour int `json:"hour"`
	Min  int `json:"minute"`
}

type tuner struct {
	Name              string               `json:"name"`
	CampaignId        string               `json:"campaignId"`
	Schedule          string               `json:"schedule"`
	DailyStartTime    timeInRequiredFormat `json:"dailyStartTime"`
	DailyEndTime      timeInRequiredFormat `json:"dailyEndTime"`
	BaselineBid       float64              `json:"baselineBid"`
	MaxBid            float64              `json:"maxBid"`
	MinBid            float64              `json:"minBid"`
	OlCutoff          float64              `json:"olCutoff"`
	MinDACutoff       float64              `json:"minDACutoff"`
	MaxDACutoff       float64              `json:"maxDACutoff"`
	DALookBackHours   float64              `json:"DALookBackHours"`
	OLLookBackMinutes float64              `json:"OLLookBackMinutes"`
	StoreIds          []string             `json:"storeIds"`
}

func (t *timeInRequiredFormat) isValid() bool {
	if t.Hour < 0 || t.Hour > 23 {
		return false
	}
	if t.Min < 0 || t.Min > 59 {
		return false
	}
	return true
}

func (t *tuner) isValid() int {
	if !isValidCronExpression(t.Schedule) {
		return 0
	}
	if !t.DailyStartTime.isValid() || !t.DailyEndTime.isValid() {
		return 1
	}
	return 2
}

func isValidCronExpression(expression string) bool {
	_, err := cron.ParseStandard(expression)
	return err == nil
}

var tuners []tuner

func getTuners(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, tuners)
}

func getTunerByName(name string) (*tuner, error) {
	for i, t := range tuners {
		if t.Name == name {
			return &tuners[i], nil
		}
	}

	return nil, errors.New("Tuner not found")
}

func getTuner(context *gin.Context) {
	name := context.Param("name")
	tuner, err := getTunerByName(name)

	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Tuner not found"})
	}

	context.IndentedJSON(http.StatusOK, tuner)
}

func addTuner(context *gin.Context) {
	var newTuner tuner

	if err := context.BindJSON(&newTuner); err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Tuner"})
		return
	}
	res := newTuner.isValid()
	if res == 0 {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Tuner: cron issue"})
		return
	}
	if res == 1 {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid Tuner: time issue"})
		return
	}

	tuners = append(tuners, newTuner)

	context.IndentedJSON(http.StatusCreated, newTuner)
}

func main() {
	router := gin.Default()

	router.GET("/campaign-tuners", getTuners)
	router.GET("/campaign-tuners/:name", getTuner)
	router.POST("/campaign-tuners", addTuner)

	router.Run("localhost:9090")
}
