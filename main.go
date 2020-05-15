package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

const nameQuestionId = "3"
const messageQuestionId = "2"

type FriendResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func main() {
	err := godotenv.Load(filepath.Join("./", ".env"))
	if err != nil {
		log.Fatal("Error loading .env file: ", err.Error())
	}

	arguments := os.Args[1:]

	// Only call the API if the --fetch argument is provided
	if len(arguments) > 0 && arguments[0] == "--fetch" {
		responses := getResponses()
		writeJson(responses)
		os.Exit(0)
	}

	getMessage()
}

// Get a list of the survey responses from SurveyGizmo
func getResponses() []FriendResponse {
	var result map[string]interface{}
	var FriendResponses []FriendResponse
	surveyId := os.Getenv("SURVEY_ID")
	apiToken := os.Getenv("API_TOKEN")
	apiSecret := os.Getenv("API_SECRET")

	response, err := http.Get("https://restapi.surveygizmo.com/v5/survey/" + surveyId +
		"/surveyresponse?api_token=" + apiToken +
		"&api_token_secret=" + apiSecret)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	json.Unmarshal([]byte(body), &result)

	responseData := result["data"].([]interface{})

	for _, value := range responseData {
		surveyData := value.(map[string]interface{})["survey_data"]
		message := getAnswer(surveyData, messageQuestionId)
		respondentName := getAnswer(surveyData, nameQuestionId)
		FriendResponses = append(FriendResponses, FriendResponse{
			Name:    respondentName,
			Message: message,
		})
	}

	return FriendResponses
}

// Parse the response for the actual answer
func getAnswer(surveyData interface{}, questionID string) string {
	return fmt.Sprintf("%s", surveyData.(map[string]interface{})[questionID].(map[string]interface{})["answer"])
}

// Save the responses to a JSON file
func writeJson(responses []FriendResponse) {
	jsonResponses, err := json.Marshal(responses)
	if err != nil {
		log.Fatal(err.Error())
	}

	ioutil.WriteFile(os.Getenv("JSON_PATH"), jsonResponses, 0644)
}

// Grab a random response from the JSON file and print it to STDOUT
// for cowsay (or whatever) to use
func getMessage() {
	var messages []FriendResponse

	rand.Seed(time.Now().UTC().UnixNano())

	jsonFile, err := os.Open(os.Getenv("JSON_PATH"))
	if err != nil {
		log.Fatal("Failed to open JSON file: ", err.Error())
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &messages)

	message := messages[rand.Intn(len(messages))]
	fmt.Println(message.Message, "\n", message.Name)
}
