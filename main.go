package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPass := getEnv("REDIS_PASSWORD", "")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPass,
		DB:       0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.Methods(http.MethodGet).Path("/").HandlerFunc(indexHandler)
	r.Methods(http.MethodGet).Path("/quote").HandlerFunc(quoteOfTheDayHandler(redisClient))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Starting server...")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	waitForShutdown(srv)

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(fmt.Sprintf("Welcome to go-redis-k8s!, pls /quote for get new quote!\n")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func quoteOfTheDayHandler(client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentTime := time.Now()
		date := currentTime.Format("2006-01-02")
		quote, err := client.Get(date).Result()

		if err != nil {
			resp, err := getQuoteFromAPI()
			if err != nil {
				_, err := w.Write([]byte(err.Error()))
				if err != nil {
					return
				}
			}
			quote = resp.Content
			client.Set(date, quote, 24*time.Hour)
			_, err = w.Write([]byte(quote))
			if err != nil {
				return
			}
		} else {
			_, err := w.Write([]byte(quote))
			if err != nil {
				return
			}
		}
	}
}

func getQuoteFromAPI() (*Quote, error) {
	ApiUrl := "https://api.quotable.io/random"
	resp, err := http.Get(ApiUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		quoteResp := &Quote{}
		err := json.NewDecoder(resp.Body).Decode(quoteResp)
		if err != nil {
			return nil, err
		}
		return quoteResp, nil
	} else {
		return nil, err
	}
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		return
	}

	log.Println("Shutting down")
	os.Exit(0)
}

func getEnv(key string, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value
}
