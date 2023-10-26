package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/goombaio/namegenerator"
)

type pskRespData struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Identity     string    `json:"identity"`
	CreatedAt    time.Time `json:"createdAt"`
	PreSharedKey string    `json:"preSharedKey"`
}

type goliothDevice struct {
	ProjectID   string   `json:"projectId"` // TODO: query the project id from the golioth API
	Name        string   `json:"name"`
	DeviceId    string   `json:"deviceIds"` // uuid
	hardwareIDs []string `json:"hardwareIds"`
}

// createDevice calls the golioth API to create a new device
func createGoliothDevice() (goliothDevice, error) {

	// generate device name
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	name := nameGenerator.Generate()

	// generate device id
	did := uuid.New().String()

	device := goliothDevice{
		ProjectID: projectID,
		Name:      name,
		DeviceId:  did,
	}

	body, err := json.Marshal(device)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %s", err)
	}

	// Create the API endpoint URL to create a device
	url := fmt.Sprintf("%s/v1/projects/%s/devices", baseURL, device.ProjectID)

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	// Add headers to the HTTP request
	req.Header.Set("X-API-Key", apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to create device, status code: %d, response: %s", resp.StatusCode, respBody)
	}

	type deviceData struct {
		ID          string      `json:"id"`
		HardwareIDs []string    `json:"hardwareIds"`
		Name        string      `json:"name"`
		CreatedAt   string      `json:"createdAt"`
		UpdatedAt   string      `json:"updatedAt"`
		TagIds      []string    `json:"tagIds"`
		Data        interface{} `json:"data"`
		LastReport  interface{} `json:"lastReport"`
		Status      string      `json:"status"`
		Metadata    interface{} `json:"metadata"`
		Enabled     bool        `json:"enabled"`
	}

	type Response struct {
		Data deviceData `json:"data"`
	}

	// Parse the JSON string into a Response struct
	var response Response
	err1 := json.Unmarshal([]byte(respBody), &response)
	if err1 != nil {
		log.Fatalf("Error parsing JSON: %v", err1)
		return device, err
	}

	device.DeviceId = response.Data.ID
	device.hardwareIDs = response.Data.HardwareIDs

	return device, nil
}

// createPrivateKey creates a private key for the device after the device itself has been created
func createPSK(deviceID string) (pskRespData, error) {
	type goliothPSKreq struct {
		PreSharedKey string `json:"preSharedKey"`
	}

	psk := goliothPSKreq{
		PreSharedKey: "string",
	}

	body, err := json.Marshal(psk)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %s", err)
	}

	// Create the API endpoint URL
	url := fmt.Sprintf("%s/v1/projects/%s/devices/%s/credentials", baseURL, projectID, deviceID)

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	// Add headers to the HTTP request
	req.Header.Set("X-API-Key", apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to create device psk, status code: %d, response: %s", resp.StatusCode, respBody)
	}

	type deviceData struct {
		ID           string `json:"id"`
		Type         string `json:"type"`
		Identity     string `json:"identity"`
		CreatedAt    string `json:"createdAt"`
		PreSharedKey string `json:"preSharedKey"`
	}

	type Response struct {
		Data deviceData `json:"data"`
	}

	// Parse the JSON string into a Response struct
	var response Response
	err1 := json.Unmarshal([]byte(respBody), &response)
	if err1 != nil {
		log.Fatalf("Error parsing JSON: %v", err1)
		return pskRespData{}, err
	}

	//unmarshal response
	var pskData pskRespData
	pskData.ID = response.Data.ID
	pskData.Type = response.Data.Type
	pskData.Identity = response.Data.Identity
	pskData.CreatedAt, err = time.Parse(time.RFC3339, response.Data.CreatedAt)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
		return pskRespData{}, err
	}
	pskData.PreSharedKey = response.Data.PreSharedKey

	return pskData, nil

}
