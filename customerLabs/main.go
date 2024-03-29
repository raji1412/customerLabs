package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ConvertedData struct {
	Event           string            `json:"event"`
	EventType       string            `json:"event_type"`
	AppID           string            `json:"app_id"`
	UserID          string            `json:"user_id"`
	MessageID       string            `json:"message_id"`
	PageTitle       string            `json:"page_title"`
	PageURL         string            `json:"page_url"`
	BrowserLanguage string            `json:"browser_language"`
	ScreenSize      string            `json:"screen_size"`
	Attributes      map[string]Detail `json:"attributes"`
	UserTraits      map[string]Detail `json:"traits"`
}

type Attribute struct {
	Name  string
	Key   string
	Value string
}

type Detail struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

func main() {
	// Channel to receive incoming requests
	requests := make(chan map[string]interface{})

	// Start worker pool
	go receiver(requests)

	// HTTP handler function
	http.HandleFunc("/input", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var data map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error decoding JSON: %v", err)
			return
		}

		// Send the request to the worker via the channel
		requests <- data

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Received form data successfully")
	})

	// Start HTTP server
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func receiver(requests <-chan map[string]interface{}) {
	for data := range requests {
		// Convert the data to the desired format
		convertedData := convertData(data)

		// Send the converted data to the webhook URL
		sendToWebhook(convertedData)
	}
}

func getAttributes(data map[string]interface{}) map[string]Detail {
	finalAttribute := make(map[string]Detail)
	i := 1
	for {
		nameKey := fmt.Sprintf("atrk%d", i)
		valueKey := fmt.Sprintf("atrv%d", i)
		typeKey := fmt.Sprintf("atrt%d", i)

		name, nameOk := data[nameKey].(string)
		value, valueOk := data[valueKey].(string)
		typeVal, typeOk := data[typeKey].(string)

		if nameOk && valueOk && typeOk {
			finalAttribute[name] = Detail{
				Value: value,
				Type:  typeVal,
			}
			i++
		} else {
			break
		}
	}
	return finalAttribute
}

func getUserTraits(data map[string]interface{}) map[string]Detail {
	finalAttribute := make(map[string]Detail)
	i := 1
	for {
		nameKey := fmt.Sprintf("uatrk%d", i)
		valueKey := fmt.Sprintf("uatrv%d", i)
		typeKey := fmt.Sprintf("uatrt%d", i)

		name, nameOk := data[nameKey].(string)
		value, valueOk := data[valueKey].(string)
		typeVal, typeOk := data[typeKey].(string)

		if nameOk && valueOk && typeOk {
			finalAttribute[name] = Detail{
				Value: value,
				Type:  typeVal,
			}
			i++
		} else {
			break
		}
	}
	return finalAttribute
}

func convertData(data map[string]interface{}) *ConvertedData {
	attributes := make(map[string]Detail)
	usertraits := make(map[string]Detail)

	// Convert attributes
	attributes = getAttributes(data)
	usertraits = getUserTraits(data)

	// Construct converted data
	convertedData := &ConvertedData{
		Event:           data["ev"].(string),
		EventType:       data["et"].(string),
		AppID:           data["id"].(string),
		UserID:          data["uid"].(string),
		MessageID:       data["mid"].(string),
		PageTitle:       data["t"].(string),
		PageURL:         data["p"].(string),
		BrowserLanguage: data["l"].(string),
		ScreenSize:      data["sc"].(string),
		Attributes:      attributes,
		UserTraits:      usertraits, // Initialize empty map for user traits
	}

	return convertedData
}

func sendToWebhook(data *ConvertedData) {
	webhookURL := "https://webhook.site/"

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}
	fmt.Println("data : ", string(jsonData))

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending data to webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Data sent to webhook successfully")
}
