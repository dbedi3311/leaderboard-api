package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/leaderboard-api/docs"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Swagger Leaderboard API
// @version 1.0
// @description This is an API to store scores and efficiently query for rankings
// @termsOfService http://swagger.io/terms/

const port string = ":8080"

var ctx = context.Background()
var client *redis.Client

type ScoreSubmission struct {
	Username string  `json:"username"`
	Score    float64 `json:"score"`
}

func init() {
	// make sure Redis client is running
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
}

// HomeHandler godoc
// @Summary Home endpoint
// @Description Home endpoint returns an up and running message
// @Tags home
// @Produce  plain
// @Success 200 {string} string "Up and running!\n"
// @Router / [get]
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Up and running!\n"))
}

// submitScoreHandler godoc
// @Summary Submit score
// @Description Submit a user score
// @Tags scores
// @Accept  json
// @Produce  json
// @Param scoreSubmission body ScoreSubmission true "Score Submission"
// @Success 201
// @Router /submit-score [post]
func submitScoreHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST /submit-score")
	var scoreSubmission ScoreSubmission

	err := json.NewDecoder(r.Body).Decode(&scoreSubmission)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	username := scoreSubmission.Username
	score := scoreSubmission.Score

	// Submit score to Redis Sorted Set
	_, err = client.ZAdd(ctx, "leaderboard", &redis.Z{Score: score, Member: username}).Result()
	if err != nil {
		http.Error(w, "Failed to submit score", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// getLeaderboardHandler godoc
// @Summary Get leaderboard
// @Description Get the top scores from the leaderboard
// @Tags scores
// @Produce  json
// @Success 200 {array} map[string]interface{}
// @Router /leaderboard [get]
func getLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET /leaderboard")

	// Retrieve leaderboard from Redis Sorted Set
	result, err := client.ZRevRangeWithScores(ctx, "leaderboard", 0, 9).Result()
	if err != nil {
		http.Error(w, "Failed to retrieve leaderboard.", http.StatusInternalServerError)
		return
	}

	// Format the result
	var leaderboard []map[string]interface{}
	for _, member := range result {
		user := map[string]interface{}{
			"username": member.Member.(string),
			"score":    member.Score,
		}
		leaderboard = append(leaderboard, user)
	}

	// Convert to JSON and send the response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(leaderboard)
	if err != nil {
		http.Error(w, "Failed to encode leaderboard as JSON", http.StatusInternalServerError)
		return
	}
}

// getRankHandler godoc
// @Summary Get user rank
// @Description Get the rank of a user by username
// @Tags scores
// @Produce  json
// @Param username path string true "Username"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{} "error"
// @Router /rank/{username} [get]
func getRankHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	username := params["username"]

	fmt.Printf("GET /rank/%s\n", username)

	// Get rank of the user from Redis Sorted Set
	rank, err := client.ZRevRank(ctx, "leaderboard", username).Result()
	if err != nil {
		http.Error(w, "Failed to get user rank", http.StatusInternalServerError)
		return
	}

	// If the user is not found, return a JSON response with an appropriate message
	if rank == -1 {
		response := map[string]interface{}{
			"error": fmt.Sprintf("User %s not found on the leaderboard", username),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert rank to 1-based index
	rank++

	// Prepare JSON response
	response := map[string]interface{}{
		"username": username,
		"rank":     rank,
	}

	// Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", HomeHandler)

	//Store and Retrieve a Simple String in Redis.
	// err := client.Set(ctx, "foo", "bar", 0).Err()
	// if err != nil {
	// 	panic(err)
	// }

	// val, err := client.Get(ctx, "foo").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("foo", val)

	// --> after testing, foo returns bar, redis client works.

	//Define endpoints
	router.HandleFunc("/submit-score", submitScoreHandler).Methods("POST")
	router.HandleFunc("/leaderboard", getLeaderboardHandler).Methods("GET")
	router.HandleFunc("/rank/{username}", getRankHandler).Methods("GET")

	// Serve Swagger UI
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods("GET")

	// Host Server
	fmt.Println("Server Running on", port)
	log.Fatal(http.ListenAndServe(port, router))
}
