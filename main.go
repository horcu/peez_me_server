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
	"math/rand"
	"net/http"
	"os"
	"strconv"
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
	PlayersIds         [2]string       `json:"playersIds"`
	Word               string          `json:"word"`
	Definition         string          `json:"Definition"`
	MissingLetterIndex int             `json:"missingLetterIndex"`
	GameId             string          `json:"gameId"`
	RoundTime          int             `json:"roundTime"`
	PlayIndex          int             `json:"playIndex"`
	PlayerTurnId       string          `json:"playerTurnId"`
	Plays              map[string]play `json:"plays"`
	LeaderId           string          `json:"leaderId"`
	PlayDirection      string          `json:"playDirection"`
	Barriers           []string        `json:"barriers"`
	Obstacles          []string        `json:"obstacles"`
	Rewards            []string        `json:"rewards"`
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

type play struct {
	Word          string         `json:"word"`
	GameId        string         `json:"gameId"`
	UserId        string         `json:"userId"`
	TileLocations []TileLocation `json:"tileLocations"`
	Definition    string         `json:"definition"`
	PlayDirection string         `json:"playDirection"`
	PlayIndex     int            `json:"playIndex"`
}

type deleteRequest struct {
	UserId string `json:"userId"`
	GameId string `json:"gameId"`
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
	Index      int    `json:"index"`
	Letter     string `json:"letter"`
	AreaName   string `json:"areaName"`
	UserId     string `json:"userId"`
	IsSelected bool   `json:"isSelected"`
}

type Word struct {
	Word string `json:"word"`
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
	// import all words here ??

	// Prepare template for execution.
	tmpl = template.Must(template.ParseFiles("index.html"))
	data = templateData{
		Service:  service,
		Revision: revision,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/new", newGameHandler)
	mux.HandleFunc("/game/submit", nextPlayHandler)
	mux.HandleFunc("/game/delete", gameDeleteHandler)

	log.Printf("Listening on port %s", "8080")
	log.Print("Open http://localhost:8080/new in the browser", 8080)

	err := http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}

// begins a new game
func newGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/new" {
		http.NotFound(w, r)
		return
	}

	// fetch word from list of words in firebase on somewhere else loaded in memory ??
	// create an instance of a new game

	// generate the game id
	gId := uuid.NewString()

	// todo get these from the main server governing game play (AGONES ??)
	pIds := [2]string{"12345", "54321"}

	dbBaseUrl := "https://peezme-default-rtdb.firebaseio.com/"
	// init the sdk
	conf := &firebase.Config{
		ProjectID:   "peezme",
		DatabaseURL: dbBaseUrl,
	}

	app, err := firebase.NewApp(context.Background(), conf) // conf, opt)
	if err != nil {
		_ = fmt.Errorf("error initializing app: %v", err)
	}

	// connect to the client
	client, err := app.Database(context.Background())

	src := rand.NewSource(time.Now().UnixNano())
	ra := rand.New(src).Intn(15232)

	// todo get the length of the word from storage instead
	clientPath := "words/seven/" + strconv.Itoa(ra)
	wordRef := client.NewRef(clientPath)

	result := Word{}
	err = wordRef.Get(context.Background(), &result)
	if err != nil {
		_, err = fmt.Fprint(w, err)
	}

	// generate random barrier location indices
	// todo account for the home locations somehow ??
	var indices []int
	for i := 1; i < 21; i++ {
		rb := rand.New(src).Intn(169)
		indices = append(indices, rb)
	}

	areaNames := []string{
		"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a10", "a11", "a12", "a13", // "a14", "a15",
		"b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9", "b10", "b11", "b12", "b13", //"b14", "b15",
		"c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "c10", "c11", "c12", "c13", //"c14", "c15",
		"d1", "d2", "d3", "d4", "d5", "d6", "d7", "d8", "d9", "d10", "d11", "d12", "d13", // "d14", "d15",
		"e1", "e2", "e3", "e4", "e5", "e6", "e7", "e8", "e9", "e10", "e11", "e12", "e13", //"e14", "e15",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12", "f13", //"f14", "f15",
		"g1", "g2", "g3", "g4", "g5", "g6", "g7", "g8", "g9", "g10", "g11", "g12", "g13", //"g14", "g15",
		"h1", "h2", "h3", "h4", "h5", "h6", "h7", "h8", "h9", "h10", "h11", "h12", "h13", //"h14", "h15",
		"i1", "i2", "i3", "i4", "i5", "i6", "i7", "i8", "i9", "i10", "i11", "i12", "i13", //"i14", "i15",
		"j1", "j2", "j3", "j4", "j5", "j6", "j7", "j8", "j9", "j10", "j11", "j12", "j13", //"j14", "j15",
		"k1", "k2", "k3", "k4", "k5", "k6", "k7", "k8", "k9", "k10", "k11", "k12", "k13", //"k14", "k15",
		"l1", "l2", "l3", "l4", "l5", "l6", "l7", "l8", "l9", "l10", "l11", "l12", "l13", //"l14", "l15",
		"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10", "m11", "m12", "m13", //"m14", "m15",
		//"n1", "n2", "n3", "n4", "n5", "n6", "n7", "n8", "n9", "n10", "n11", "n12", "n13", "n14", "n15",
		//"o1", "o2", "o3", "o4", "o5", "o6", "o7", "o8", "o9", "o10", "o11", "o12", "o13", "o14", "o15",
	}
	var barriers []string
	var obstacles []string
	var rewards []string

	//todo change the number of each of these as the levels change

	for i := 0; i < len(indices); i++ {
		if i < 9 {
			// use 20 for barriers
			barriers = append(barriers, areaNames[indices[i]])
		} else if i < 16 {
			// use 10 for obstacles
			obstacles = append(obstacles, areaNames[indices[i]])
		} else {
			// use 5 for rewards
			rewards = append(rewards, areaNames[indices[i]])
		}
	}

	newGame := game{
		PlayersIds:         pIds,
		Word:               result.Word, // fetch from the api
		Definition:         "",
		MissingLetterIndex: 4,
		RoundTime:          120,
		PlayIndex:          0,
		Plays:              map[string]play{},
		GameId:             gId,
		PlayerTurnId:       "12345",
		PlayDirection:      "Vertical",
		Barriers:           barriers,
		Obstacles:          obstacles,
		Rewards:            rewards,
	}

	fmt.Println(newGame)

	// save game
	// Get a database reference to our game details.
	gameRef := client.NewRef("games/" + gId)

	// save details to firebase
	err = gameRef.Set(context.Background(), &newGame)
	if err != nil {
		errorResponse := ErrorResponse{Error: "invalid word"}
		_, err = fmt.Fprint(w, errorResponse)
	}

	err = json.NewEncoder(w).Encode(newGame)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// deletes an unsuccessful game
func gameDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path != "/game/delete" {
		http.NotFound(w, r)
		return
	}

	var req deleteRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

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

	// remove the game info from the database
	err = deleteGameFromDb(w, ref, req.GameId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

// handle plays within the game
func nextPlayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path != "/game/submit" {
		http.NotFound(w, r)
		return
	}

	//decode the body of the request
	var req play
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

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

	wordCheckRef := database.NewRef("words")
	// validate the request
	isAGoodWord := checkWordValidity(wordCheckRef, req.Word) //validateWord(req)

	// return a not good word response
	if !isAGoodWord {
		// remove the game info from the database
		// err := deleteGameFromDb(w, ref, req.GameId)

		// get game details

		// build the returned response
		response := wordSubmittedResponse{
			Score:              0,
			Word:               req.Word,
			Definition:         "",
			GameId:             req.GameId,
			MissingLetterIndex: len(req.Word) - 1,
			PlayDirection:      req.PlayDirection,
			PlayIndex:          req.PlayIndex,
			WordIsGood:         false,
			LeaderId:           req.UserId,
			TileLocations:      req.TileLocations,
		}

		err = json.NewEncoder(w).Encode(&response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	} else {

		// save the request as a play in the DB
		_, err = ref.Child(req.GameId).Child("plays").Push(context.Background(), req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		// get the existing game details
		var details game
		details, err := getExistingGameDetails(ref.Child(req.GameId))
		if err != nil {
			return
		}

		//update the direction
		details.PlayDirection = togglePlayDirection(details.PlayDirection)
		x := map[string]interface{}{
			"playDirection": details.PlayDirection,
		}
		err = ref.Child(req.GameId).Update(context.Background(), x)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		// increment the play index
		_playIndex := incrementPlayIndex(details)
		req.PlayIndex = _playIndex

		// get all the tile locations to return to board

		var lst []TileLocation

		// create the return tile location list
		i := 0
		for _, v := range details.Plays {
			// check if the index reps the last set of plays(tile locations)

			// if this is the last added set of tiles
			if i == len(details.Plays)-1 {

				// set the isSelected Property based on predefined rules
				v.TileLocations[len(v.TileLocations)-1].IsSelected = true
			}
			// append to return list
			lst = append(lst, v.TileLocations...)
			i++
		}

		// create an instance of a response
		response := wordSubmittedResponse{
			Score:              score,
			Word:               req.Word,
			Definition:         "",
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

func deleteGameFromDb(w http.ResponseWriter, ref *db.Ref, gameId string) error {
	err := ref.Child(gameId).Delete(context.Background())
	if err != nil {
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	return err
}

func checkWordValidity(ref *db.Ref, word string) bool {
	// filter map
	var NumberToWord = map[int]string{
		1:  "one",
		2:  "two",
		3:  "three",
		4:  "four",
		5:  "five",
		6:  "six",
		7:  "seven",
		8:  "eight",
		9:  "nine",
		10: "ten",
		11: "eleven",
		12: "twelve",
		13: "thirteen",
		14: "fourteen",
		15: "fifteen",
	}

	//count letters
	count := len(word)
	url := NumberToWord[count]
	// use variable to determine list to search

	val, err := ref.Child(url).OrderByChild("word").EqualTo(word).GetOrdered(context.Background())
	if err != nil {
		return false
	}
	return len(val) > 0
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

func validateWord(req play) (bool, string) {

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
