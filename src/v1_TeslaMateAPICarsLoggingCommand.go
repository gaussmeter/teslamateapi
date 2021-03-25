package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// TeslaMateAPICarsLoggingCommandV1 func
func TeslaMateAPICarsLoggingCommandV1(c *gin.Context) {

	// creating required vars
	var jsonData map[string]interface{}
	var err error

	// check if commands are enabled.. if not we need to abort
	if getEnvAsBool("ENABLE_COMMANDS", false) == false {
		log.Println("[warning] TeslaMateAPICarsLoggingCommandV1 ENABLE_COMMANDS is not true.. returning 403 forbidden.")
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to access logging commands"})
		return
	}

	// if request method is GET return list of commands
	if c.Request.Method == http.MethodGet {
		c.JSON(http.StatusOK, gin.H{"enabled_commands": allowList})
		return
	}

	// authentication for the endpoint
	validToken, errorMessage := validateAuthToken(c)
	if !validToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
		return
	}

	// getting CarID param from URL
	ParamCarID := c.Param("CarID")
	var CarID int
	if ParamCarID != "" {
		CarID = convertStringToInteger(ParamCarID)
	}

	// validating that CarID is not zero
	if CarID == 0 {
		log.Println("[error] TeslaMateAPICarsLoggingCommandV1 CarID is invalid (zero)!")
		c.JSON(http.StatusBadRequest, gin.H{"error": "CarID invalid"})
		return
	}

	// getting request body to pass to Tesla
	reqBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("[error] TeslaMateAPICarsLoggingCommandV1 error in first ioutil.ReadAll", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	// getting :Command
	command := ("/logging/" + c.Param("Command"))
	log.Println("[debug] TeslaMateAPICarsLoggingCommandV1 command received:", command)

	if !checkArrayContainsString(allowList, command) {
		log.Print("[warning] TeslaMateAPICarsCommandV1 command: " + command + " not allowed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	client := &http.Client{}
	putURL := getEnv("TESLAMATE_URL", "http://teslamate:4000") + "/api/car/" + ParamCarID + command
	req, _ := http.NewRequest(http.MethodPut, putURL, strings.NewReader(string(reqBody)))
	req.Header.Set("User-Agent", "TeslaMateApi/"+apiVersion+" https://github.com/tobiasehlert/teslamateapi")
	resp, err := client.Do(req)

	// check response error
	if err != nil {
		log.Println("[error] TeslaMateAPICarsLoggingCommandV1 error in http request to http://teslamate:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	defer resp.Body.Close()
	defer client.CloseIdleConnections()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("[error] TeslaMateAPICarsLoggingCommandV1 error in second ioutil.ReadAll:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}
	json.Unmarshal([]byte(respBody), &jsonData)

	// print to log about request
	if gin.IsDebugging() {
		log.Println("[debug] TeslaMateAPICarsLoggingCommandV1 " + c.Request.RequestURI + " returned data:")
		js, _ := json.Marshal(jsonData)
		log.Printf("[debug] %s\n", js)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		log.Println("[info] TeslaMateAPICarsLoggingCommandV1 " + c.Request.RequestURI + " executed successful.")
	} else {
		log.Println("[error] TeslaMateAPICarsLoggingCommandV1 " + c.Request.RequestURI + " error in execution!")
	}
	c.JSON(resp.StatusCode, jsonData)
}
