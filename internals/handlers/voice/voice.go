package handlers

import (
	"fmt"
	"strings"

	"github.com/KashyretsIvanna/voice-balance/internals/services"
	"github.com/gofiber/fiber/v2"
)

// TranscribeAudio godoc
// @Summary      Transcribe audio to text
// @Description  Receives an audio file and transcribes it to text using Google Cloud Speech-to-Text
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Tags         transcription
// @Accept       audio/wav
// @Produce      json
// @Param        file  formData  file  true  "Audio file to transcribe"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/voice [post]
func TranscribeAudio(c *fiber.Ctx) error {
	// Parse the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to get file: Please upload a valid file.",
		})
	}

	// Attempt to parse the file and transcribe the audio
	transcription, err := services.ParseText(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to transcribe the uploaded file: %v", err),
		})
	}

	// Convert the transcribed text to lowercase
	textCommand := strings.ToLower(transcription)

	// Use the AskAi service to interpret the text command
	err, res := services.AskAi(textCommand)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("AI processing error: %v", err),
		})
	}

	// Return the successfully processed action
	return c.JSON(fiber.Map{
		"status": "success",
		"action": res,
	})
}