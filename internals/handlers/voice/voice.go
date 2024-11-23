package handlers

import (
	"fmt"

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
	//Parse the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed ton get file",
		})
	}


	fmt.Print("before")
	transcription, err := services.ParseText(file)
	fmt.Print(transcription)
	fmt.Print(err)



	if err != nil {
		fmt.Print(err)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	action, err := services.GetActionFromVoice(transcription)
	if err != nil {
		fmt.Print(err)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"action": action,
	})
}
