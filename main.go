package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	tracker "github.com/kevingentile/fortnite-tracker"
)

func main() {
	log.Println("obs-fortnite starting on port: " + os.Getenv("PORT"))

	trackerAPILimiter := time.Tick(time.Second * 3)

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.Static("/assets", "./assets")

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/obs/fortnite")
	})

	router.GET("/obs", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/obs/fortnite")
	})

	router.GET("/obs/fortnite", func(c *gin.Context) {
		c.HTML(http.StatusOK, "fortnite-stats-form.tmpl.html", nil)
	})

	router.GET("/fortnite/:platform/:username", func(c *gin.Context) {
		c.File("pages/fortnite-stats-widget.html")
	})

	router.GET("/obs/fortnite/:platform/:username", rateLimit(func(c *gin.Context) {
		platform := c.Param("platform")
		username := c.Param("username")
		log.Println("Fornite Tracker lookup for: " + username + " " + platform)

		token := os.Getenv("FORTNITE_TRACKER_TOKEN")
		log.Println("token: " + token)

		profile, err := tracker.GetProfile(platform, username, token)
		if err != nil {
			handleTrackerError(err, c)
		}

		kills, err := profile.GetKills()
		if err != nil {
			handleTrackerError(err, c)
			return
		}

		wins, err := profile.GetWins()
		if err != nil {
			handleTrackerError(err, c)
			return
		}

		kdr, err := profile.GetKDR()
		if err != nil {
			handleTrackerError(err, c)
			return
		}

		fortniteData := &FortniteData{
			Kills: kills,
			Wins:  wins,
			KDR:   kdr,
		}

		log.Println(fortniteData)

		c.JSON(http.StatusOK, fortniteData)

	}, &trackerAPILimiter))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))
}

func rateLimit(handle gin.HandlerFunc, limiter *<-chan time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		<-*limiter
		handle(c)
	}
}

func handleTrackerError(err error, c *gin.Context) {
	if err != nil {
		log.Println("Error proxying fornite stat request: ", err)
		data := FortniteData{
			Wins:  -1,
			KDR:   -1,
			Kills: -1,
		}
		c.JSON(http.StatusInternalServerError, data)
	}
}

type FortniteData struct {
	Wins  int     `json:"wins"`
	KDR   float64 `json:"kdr"`
	Kills int     `json:"kills"`
}
