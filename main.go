package main

import (
	"context"
	"encoding/json"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	fmt "fmt"
	"github.com/google/uuid"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// templateData provides template parameters.
type templateData struct {
	Service  string
	Revision string
}

// Variables used to generate the HTML page.
var (
	data templateData
	tmpl *template.Template
)

type game struct {
	PlayersIds         [2]string                       `json:"playersIds"`
	Word               string                          `json:"word"`
	Definition         string                          `json:"Definition"`
	MissingLetterIndex int                             `json:"missingLetterIndex"`
	GameId             string                          `json:"gameId"`
	RoundTime          int                             `json:"roundTime"`
	PlayIndex          int                             `json:"playIndex"`
	PlayerTurnId       string                          `json:"playerTurnId"`
	Plays              map[string]wordSubmittedRequest `json:"plays"`
	LeaderId           string                          `json:"leaderId"`
	PlayDirection      string                          `json:"playDirection"`
}

type wordSubmittedResponse struct {
	Score              int            `json:"score"`
	Word               string         `json:"word"`
	GameId             string         `json:"gameId"`
	MissingLetterIndex int            `json:"missingLetterIndex"`
	PlayIndex          int            `json:"playIndex"`
	PlayerTurnId       string         `json:"playerTurnId"`
	WordIsGood         bool           `json:"wordIsGood"`
	PlayDirection      string         `json:"playDirection"`
	LeaderId           string         `json:"leaderId"`
	TileLocations      []TileLocation `json:"tileLocations"`
	Definition         string         `json:"Definition"`
}

type wordSubmittedRequest struct {
	Word          string         `json:"word"`
	GameId        string         `json:"gameId"`
	UserId        string         `json:"userId"`
	TileLocations []TileLocation `json:"tileLocations"`
	Definition    string         `json:"definition"`
	PlayDirection string         `json:"playDirection"`
	PlayIndex     int            `json:"playIndex"`
}

type WordDefinition struct {
	Word       string `json:"word"`
	Valid      bool   `json:"valid"`
	Definition string `json:"Definition"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TileLocation struct {
	Index    int    `json:"index"`
	Letter   string `json:"letter"`
	AreaName string `json:"areaName"`
	UserId   string `json:"userId"`
}

func main() {
	// Initialize template parameters.
	service := os.Getenv("K_SERVICE")
	if service == "" {
		service = "???"
	}

	revision := os.Getenv("K_REVISION")
	if revision == "" {
		revision = "???"
	}

	// Prepare template for execution.
	tmpl = template.Must(template.ParseFiles("index.html"))
	data = templateData{
		Service:  service,
		Revision: revision,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/new", newGameHandler)
	mux.HandleFunc("/game/submit", nextPlayHandler)

	log.Printf("Listening on port %s", "8080")
	log.Print("Open http://localhost:8080/new in the browser", 8080)

	err := http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}

func newGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/new" {
		http.NotFound(w, r)
		return
	}

	// fetch word from list of words in firebase on somewhere else loaded in memory ??
	// create an instance of a new game

	// generate the game id
	gId := uuid.NewString()
	pIds := [2]string{"12345", "54321"}

	newGame := game{
		PlayersIds:         pIds,
		Word:               "abeam", // fetch from the api
		Definition:         "on a line at right angles to a ship's or an aircraft's length",
		MissingLetterIndex: 4,
		RoundTime:          120,
		PlayIndex:          0,
		GameId:             gId,
		PlayerTurnId:       "12345",
		PlayDirection:      "Vertical",
	}

	fmt.Println(newGame)

	// init the sdk
	conf := &firebase.Config{
		ProjectID:   "peezme",
		DatabaseURL: "https://peezme-default-rtdb.firebaseio.com/",
	}

	app, err := firebase.NewApp(context.Background(), conf) // conf, opt)
	if err != nil {
		_ = fmt.Errorf("error initializing app: %v", err)
	}

	// connect to the client
	client, err := app.Database(context.Background())

	// save game
	// Get a database reference to our game details.
	ref := client.NewRef("games/" + gId)

	// save details to firebase
	err = ref.Set(context.Background(), &newGame)
	if err != nil {
		errorResponse := ErrorResponse{Error: "invalid word"}
		_, err = fmt.Fprint(w, errorResponse)
	}

	err = json.NewEncoder(w).Encode(newGame)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func nextPlayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path != "/game/submit" {
		http.NotFound(w, r)
		return
	}

	var req wordSubmittedRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// validate the request
	isAGoodWord, def := validateWord(req)

	// All good  so set the definition !!
	req.Definition = def

	// score the word todo complete function
	score := scoreWord(req.Word)

	// create config for database access
	conf := &firebase.Config{
		ProjectID:   "peezme",
		DatabaseURL: "https://peezme-default-rtdb.firebaseio.com/",
	}

	// todo make this central
	// new firebase app instance
	app, err := firebase.NewApp(context.Background(), conf)
	if err != nil {
		_ = fmt.Errorf("error initializing firebase database app: %v", err)
	}

	// connect to the database
	database, err := app.Database(context.Background())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Get a database reference to our game details.
	ref := database.NewRef("games/")

	// return a not good word response
	if !isAGoodWord {
		// remove the game info from the database
		err := ref.Child(req.GameId).Delete(context.Background())
		if err != nil {
			err = json.NewEncoder(w).Encode(err)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}

		// build the returned response
		response := wordSubmittedResponse{
			Score:              0,
			Word:               req.Word,
			Definition:         "",
			GameId:             req.GameId,
			MissingLetterIndex: 0,
			PlayDirection:      req.PlayDirection,
			PlayIndex:          req.PlayIndex,
			WordIsGood:         false,
			LeaderId:           "12345",
			TileLocations:      []TileLocation{},
		}

		err = json.NewEncoder(w).Encode(&response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	} else {

		// save the request as a play in the DB
		_, err = ref.Child(req.GameId).Child("Plays").Push(context.Background(), req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		// get the existing game details
		var details game
		details, err = getExistingGameDetails(ref.Child(req.GameId))

		//update the direction
		details.PlayDirection = togglePlayDirection(details.PlayDirection)
		err := ref.Child(req.GameId).Set(context.Background(), details)

		if err != nil {
			return
		}
		// increment the play index
		_playIndex := incrementPlayIndex(details)
		req.PlayIndex = _playIndex

		// get all the tile locations to return to board
		var lst []TileLocation
		for _, v := range details.Plays {
			//fmt.Println(i, v)
			lst = append(lst, v.TileLocations...)
		}

		// create an instance of a response
		response := wordSubmittedResponse{
			Score:              score,
			Word:               req.Word,
			Definition:         def,
			GameId:             req.GameId,
			MissingLetterIndex: getNextMissingLetterIndex(req.Word, details), // todo make this more intelligent
			PlayDirection:      togglePlayDirection(req.PlayDirection),
			PlayIndex:          _playIndex,
			WordIsGood:         true,
			LeaderId:           details.LeaderId,
			TileLocations:      lst,
		}

		// return the response
		err = json.NewEncoder(w).Encode(&response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}

func togglePlayDirection(direction string) string {
	if direction == "Horizontal" {
		return "Vertical"
	}
	return "Horizontal"
}

// todo complete function
func getNextMissingLetterIndex(word string, details game) int {
	return 2
}

func incrementPlayIndex(details game) int {
	return details.PlayIndex + 1
}

// todo return details from db
func getExistingGameDetails(ref *db.Ref) (game, error) {
	g := game{}
	err := ref.Get(context.Background(), &g)
	if err != nil {
		return game{}, err
	}
	return g, nil
}

const verificationUrl = "https://api.api-ninjas.com/v1/dictionary?word="

func validateWord(req wordSubmittedRequest) (bool, string) {

	isNotGood := req.Word == ""
	if isNotGood {
		return false, ""
	}

	verificationRequest, err := http.NewRequest(http.MethodGet, verificationUrl+req.Word, nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return false, ""
	}

	verificationRequest.Header.Set("Content-Type", "application/json")
	verificationRequest.Header.Set("X-Api-Key", "/cdoEbYKFqd7VxIhm7tyew==DWjL80jNVgehzngU")

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(verificationRequest)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return false, ""
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		return false, ""
	}

	wd := WordDefinition{}

	err = json.Unmarshal(resBody, &wd)
	if err != nil {
		fmt.Printf("client: could not decode response body: %s\n", err)
		return false, ""
	}

	return wd.Valid, wd.Definition
}

func scoreWord(word string) int {
	return 10
}
