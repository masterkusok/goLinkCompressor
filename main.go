package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/http"
	"strings"
)

const (
	AVAILABLE   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	TOKENLENGTH = 6
)

var lastToken = []byte("aaaaaa")
var availableChars = []byte(AVAILABLE)
var urls = make(map[string]string)

func generateToken() (string, error) {
	newToken := make([]byte, len(lastToken))
	copy(newToken, lastToken)
	increaseNext := false
	for i := TOKENLENGTH - 1; i >= 0; i-- {
		index := strings.IndexByte(AVAILABLE, newToken[i])
		if index != len(availableChars)-1 {
			newToken[i] = availableChars[index+1]

			if !increaseNext {
				lastToken = newToken
				return string(newToken), nil
			}

			if index+2 < len(availableChars) {
				newToken[i] = availableChars[index+1]
				lastToken = newToken
				return string(newToken), nil
			}
		}
		newToken[i] = availableChars[0]
		increaseNext = true

	}
	return "", fmt.Errorf("error during generating new token")
}

// CreateCompressedUrlHandler
//
//	 JSON SCHEMA:
//		"url": string
func CreateCompressedUrlHandler(c *fiber.Ctx) error {
	jsonText := c.Body()
	dict := map[string]string{}
	err := json.Unmarshal(jsonText, &dict)
	if err != nil {
		return c.SendStatus(http.StatusBadRequest)
	}
	// Check if compressed url already exists
	if _, ok := urls[dict["url"]]; ok {
		return c.SendStatus(http.StatusBadRequest)
	}
	token, err := generateToken()
	if err != nil {
		return c.SendStatus(http.StatusInternalServerError)
	}

	if !(strings.Contains(dict["url"], "http://") || strings.Contains(dict["url"], "https://")) {
		dict["url"] = "https://" + dict["url"]
	}
	urls[token] = dict["url"]
	err = c.SendStatus(http.StatusCreated)
	return c.Send([]byte(token))
}

func CompressedUrlHandler(c *fiber.Ctx) error {
	token := c.Params("token")

	if len(token) != 6 {
		return c.SendStatus(http.StatusBadRequest)
	}

	url, ok := urls[token]
	if !ok {
		return c.SendStatus(http.StatusNotFound)
	}
	return c.Redirect(url, http.StatusSeeOther)
}

func main() {
	app := fiber.New()

	app.Add(http.MethodPost, "/api/v1/url", CreateCompressedUrlHandler)
	app.Add(http.MethodGet, "/:token", CompressedUrlHandler)

	err := app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
