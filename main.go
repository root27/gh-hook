package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"io"
	"log"
	"net/http"
	"os"
)

func computeHMAC256(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

type SlackNotifier struct {
	Token   string
	Channel string
}

type Event any

type StarEvent struct {
	Sender string
	Repo   string
}

func (s *SlackNotifier) Notify(eventType string, event StarEvent) error {

	text := ""

	if eventType == "created" {
		text = event.Sender + " starred " + event.Repo

	} else {
		text = event.Sender + " unstarred " + event.Repo

	}

	api := slack.New(s.Token)

	_, _, err := api.PostMessage(s.Channel, slack.MsgOptionText(text, false))

	if err != nil {
		log.Println("Error sending message to Slack")

		return err
	}

	return nil
}

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")

	}

	slack := SlackNotifier{
		Token:   os.Getenv("SLACK_TOKEN"),
		Channel: os.Getenv("SLACK_CHANNEL"),
	}

	secret := os.Getenv("WEBHOOK_SECRET")

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		reqBody, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		defer r.Body.Close()

		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" {
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}
		expectedSignature := "sha256=" + computeHMAC256(reqBody, secret)

		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			http.Error(w, "Invalid signature", http.StatusForbidden)
			return
		}

		var event Event

		err = json.Unmarshal(reqBody, &event)

		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}

		// sender -> login (who starred the repo)

		//repository url -> html_url

		starEvent := StarEvent{
			Sender: event.(map[string]any)["sender"].(map[string]any)["login"].(string),
			Repo:   event.(map[string]any)["repository"].(map[string]any)["html_url"].(string),
		}

		switch event.(map[string]any)["action"] {
		case "created":
			log.Println("New star created")

			err := slack.Notify("created", starEvent)

			if err != nil {
				log.Println("Error sending notification to Slack")
			}
		case "deleted":
			log.Println("Star deleted")
			err := slack.Notify("deleted", starEvent)

			if err != nil {
				log.Println("Error sending notification to Slack")
			}

		}

		log.Println("Notification sent to Slack")
	})

	log.Println("Starting server on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))

}
