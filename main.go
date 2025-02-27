package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/joho/godotenv"
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

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")

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

		var event any

		err = json.Unmarshal(reqBody, &event)

		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}

		log.Printf("Event received: %v\n", event)

	})

	log.Println("Starting server on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))

}
