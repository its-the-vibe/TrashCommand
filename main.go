package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/slack-go/slack"
)

// ReactionEvent represents the structure of a Slack reaction event
type ReactionEvent struct {
	Token          string `json:"token"`
	TeamID         string `json:"team_id"`
	APIAppID       string `json:"api_app_id"`
	Type           string `json:"type"`
	Event          Event  `json:"event"`
	Authorizations []Auth `json:"authorizations"`
}

// Event represents the nested event object
type Event struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Reaction string `json:"reaction"`
	Item     Item   `json:"item"`
	ItemUser string `json:"item_user"`
	EventTS  string `json:"event_ts"`
}

// Item represents the item that was reacted to
type Item struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	TS      string `json:"ts"`
}

// Auth represents authorization information
type Auth struct {
	UserID string `json:"user_id"`
	IsBot  bool   `json:"is_bot"`
}

// Config holds the application configuration
type Config struct {
	SlackBotToken        string
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	RedisChannel         string
	TimeBombRedisChannel string
	TimeBombTTLSeconds   int
}

func main() {
	// Load configuration from environment variables
	config := Config{
		SlackBotToken:        getEnv("SLACK_BOT_TOKEN", ""),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		RedisDB:              0,
		RedisChannel:         getEnv("REDIS_CHANNEL", "slack-relay-reaction-added"),
		TimeBombRedisChannel: getEnv("TIMEBOMB_REDIS_CHANNEL", "timebomb-messages"),
		TimeBombTTLSeconds:   getEnvInt("TIMEBOMB_TTL_SECONDS", 5),
	}

	if config.SlackBotToken == "" {
		log.Fatal("SLACK_BOT_TOKEN environment variable is required")
	}

	// Create Slack client
	slackClient := slack.New(config.SlackBotToken)

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Connected to Redis at %s", config.RedisAddr)

	// Subscribe to Redis channel
	pubsub := redisClient.Subscribe(ctx, config.RedisChannel)
	defer pubsub.Close()

	log.Printf("Subscribed to Redis channel: %s", config.RedisChannel)
	log.Println("Waiting for reaction events...")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Listen for messages
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, exiting")
			return
		case msg, ok := <-ch:
			if !ok {
				log.Println("Channel closed, exiting")
				return
			}
			handleMessage(msg.Payload, slackClient, redisClient, &config)
		}
	}
}

// TimeBombMessage represents the message structure to send to TimeBomb
type TimeBombMessage struct {
	Channel   string `json:"channel"`
	Timestamp string `json:"ts"`
	TTL       int    `json:"ttl"`
}

func handleMessage(payload string, slackClient *slack.Client, redisClient *redis.Client, config *Config) {
	var event ReactionEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("Error parsing payload: %v", err)
		return
	}

	// Check if this is a reaction_added event
	if event.Event.Type != "reaction_added" {
		log.Printf("Skipping non-reaction event: %s", event.Event.Type)
		return
	}

	// Check if the item is a message
	if event.Event.Item.Type != "message" {
		log.Printf("Skipping non-message item: %s", event.Event.Item.Type)
		return
	}

	// Check if the user who reacted is not a bot
	if isBot(event) {
		log.Printf("Skipping bot user reaction")
		return
	}

	// Handle wastebasket reaction - delete message immediately
	if event.Event.Reaction == "wastebasket" {
		deleteMessage(event, slackClient)
		return
	}

	// Handle bomb reaction - publish to TimeBomb
	if event.Event.Reaction == "bomb" {
		publishToTimeBomb(event, redisClient, config)
		return
	}

	log.Printf("Skipping unsupported reaction: %s", event.Event.Reaction)
}

// deleteMessage deletes a Slack message immediately
func deleteMessage(event ReactionEvent, slackClient *slack.Client) {
	channel := event.Event.Item.Channel
	timestamp := event.Event.Item.TS

	log.Printf("Deleting message in channel %s with timestamp %s", channel, timestamp)

	_, _, err := slackClient.DeleteMessage(channel, timestamp)
	if err != nil {
		log.Printf("Error deleting message: %v", err)
		return
	}

	log.Printf("Successfully deleted message in channel %s", channel)
}

// publishToTimeBomb publishes a message to the TimeBomb Redis channel
func publishToTimeBomb(event ReactionEvent, redisClient *redis.Client, config *Config) {
	channel := event.Event.Item.Channel
	timestamp := event.Event.Item.TS

	message := TimeBombMessage{
		Channel:   channel,
		Timestamp: timestamp,
		TTL:       config.TimeBombTTLSeconds,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling TimeBomb message: %v", err)
		return
	}

	ctx := context.Background()
	err = redisClient.Publish(ctx, config.TimeBombRedisChannel, string(payload)).Err()
	if err != nil {
		log.Printf("Error publishing to TimeBomb: %v", err)
		return
	}

	log.Printf("Published message to TimeBomb: channel=%s, ts=%s, ttl=%ds", channel, timestamp, config.TimeBombTTLSeconds)
}

// isBot checks if the user who reacted is a bot
func isBot(event ReactionEvent) bool {
	// Check authorizations to see if the user is a bot
	for _, auth := range event.Authorizations {
		if auth.UserID == event.Event.User && auth.IsBot {
			return true
		}
	}
	return false
}

// getEnv retrieves an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt retrieves an environment variable as an integer with a default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue := 0
	_, err := fmt.Sscanf(value, "%d", &intValue)
	if err != nil {
		log.Printf("Invalid integer value for %s: %s, using default %d", key, value, defaultValue)
		return defaultValue
	}
	return intValue
}
