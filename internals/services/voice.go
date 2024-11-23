package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"regexp"
	"strings"
	"cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb" // Updated import for speech types
	"github.com/KashyretsIvanna/voice-balance/config"
	"google.golang.org/api/option"
)

func ParseText(file *multipart.FileHeader) (string, error) {

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err

	}
	defer src.Close()

	// Create a temporary file to store the audio content
	tempFile, err := os.CreateTemp("./", "audio-*")
	if err != nil {
		return "", err

	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy uploaded file content to the temp file
	if _, err := io.Copy(tempFile, src); err != nil {
		return "", err

	}

	// Initialize Google Cloud Speech client with credentials
	ctx := context.Background()
	client, err := speech.NewClient(ctx, option.WithCredentialsFile(config.Config("CLOUD_JSON_PATH")))
	if err != nil {
		return "", err

	}
	
	defer client.Close()

	fmt.Print(tempFile.Name())
	// Read the audio file data
	audioData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", err

	}

	// Configure the recognition request
	req := &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16, // Set based on audio format
			SampleRateHertz: 48000,                               // Adjust sample rate as needed
			LanguageCode:    "uk-UA",                             // Language code for transcription
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{
				Content: audioData,
			},
		},
	}

	// Perform transcription
	resp, err := client.Recognize(ctx, req)
	if err != nil {
		return "", err

	}

	fmt.Print(resp.Results)

	// Collect the transcription result
	transcription := ""
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			transcription += alt.Transcript + " "
		}
	}

	return transcription, nil

}

// Функція для обробки голосових команд
func GetActionFromVoice(command string) (map[string]interface{}, error) {
	// Перетворюємо команду в нижній регістр для полегшення порівнянь
	command = strings.ToLower(command)

	// Обробка запитів на статистику витрат і доходів
	if includes([]string{"статистика", "статистику"}, command) {
		return handleShowStatistics(command), nil
	} else if includes([]string{"додай витрату", "додай витрати"}, command) {
		return handleAddExpense(command), nil
	} else if includes([]string{"додай дохід"}, command) {
		return handleAddIncome(command), nil
	} else if includes([]string{"нагадай"}, command) {
		return handleAddReminder(command), nil
	} else {
		return map[string]interface{}{}, fmt.Errorf("Error getting type of action")
	}
}

func includes(slice []string, str string) bool {
	// Перевіряємо, чи містить рядок `str` будь-який елемент із `slice`

	for _, item := range slice {
		if strings.Contains(str, item) {
			return true
		}
	}
	return false
}

func handleAddExpense(command string) map[string]interface{} {
	// Define regex patterns to extract category and amount
	// amountPattern := `.*?\s+([а-яА-ЯіІїЇєЄґҐ\s]+)\s*грн.*?`
	amountPattern := `.*?\s+(\d+(\.\d+)?)\s*грн.*?`

	categoryPattern := `.*?на\s+([а-яА-ЯіІїЇєЄґҐ]+).*?`
	// Use regex to extract amount and category
	amountRegex := regexp.MustCompile(amountPattern)
	categoryRegex := regexp.MustCompile(categoryPattern)

	// Find the amount and category in the command
	amountMatches := amountRegex.FindStringSubmatch(command)
	categoryMatches := categoryRegex.FindStringSubmatch(command)

	fmt.Print(amountMatches)
	fmt.Print(categoryMatches)

	// Default category if not found
	category := "не вказано"
	if len(categoryMatches) > 0 {
		category = strings.TrimSpace(categoryMatches[1])
	}

	// Default amount if not found
	amount := "нуль"
	if len(amountMatches) > 0 {
		amount = amountMatches[1] // Extracts the amount from the match
	}

	// Print the parsed results for debugging
	fmt.Printf("Додано витрату: %s на категорію %s\n", amount, category)

	// Return the parsed data in a map
	result := map[string]interface{}{
		"amount":   amount,
		"category": category,
		"type":     "витрати",
	}
	return result
}

// Функція для додавання доходу
func handleAddIncome(command string) map[string]interface{} {
	// Регулярні вирази для вилучення суми та категорії
	amountPattern := `.*?\s+(\d+(\.\d+)?)\s*за.*?`
	categoryPattern := `.*?\s+за\s+категорією\s+([а-яА-ЯіІїЇєЄґҐ]+).*?`

	amountRegex := regexp.MustCompile(amountPattern)
	categoryRegex := regexp.MustCompile(categoryPattern)

	// Ініціалізуємо значення
	amount := "0"
	category := "загальна"

	// Шукаємо суму в команді
	amountMatches := amountRegex.FindStringSubmatch(command)
	if len(amountMatches) > 1 {
		amount = amountMatches[1] // Наприклад: "два" -> 2
	}

	// Шукаємо категорію в команді
	categoryMatches := categoryRegex.FindStringSubmatch(command)
	if len(categoryMatches) > 1 {
		category = categoryMatches[1] // Наприклад: "зарплата"
	}

	result := map[string]interface{}{
		"amount":   amount,
		"category": category,
		"type":     "доходи",
	}
	return result
}

// Функція для додавання нагадування
func handleAddReminder(command string) map[string]interface{} {
	input := "Нагадай оплатити рахунок за електроенергію"

	// Видалення слова "нагадай" (незалежно від регістру)
	reminder := strings.Replace(strings.ToLower(input), "нагадай", "", 1)
	reminder = strings.TrimSpace(reminder) // Видалення зайвих пробілів

	result := map[string]interface{}{
		"category": reminder,
		"type":     "нагадування",
	}
	return result

}

// Функція для показу статистики витрат та доходів
func handleShowStatistics(command string) map[string]interface{} {
	var category string
	if strings.Contains(command, "доход") {
		category = "доходи"
	}

	if strings.Contains(command, "витрат") {
		category = "витрати"
	}

	// Якщо команда вимагає статистику за місяць або тиждень
	if strings.Contains(command, "місяць") {
		// Логіка показу статистики за місяць (псевдокод)
		// Припустимо, що ми отримуємо статистику за поточний місяць

		result := map[string]interface{}{
			"category": category,
			"range":    "місяць",
			"type":     "статистика",
		}
		return result
		// Тут можна додавати реальний код для отримання статистики з бази
	} else if strings.Contains(command, "тиждень") {
		// Логіка показу статистики за тиждень
		fmt.Println("Показано статистику за поточний тиждень")
		// Тут можна додавати реальний код для отримання статистики з бази
		result := map[string]interface{}{
			"category": category,
			"range":    "тиждень",
			"type":     "статистика",
		}
		return result
	} else if strings.Contains(command, "день") {
		// Логіка показу статистики за тиждень
		fmt.Println("Показано статистику за поточний тиждень")
		// Тут можна додавати реальний код для отримання статистики з бази
		result := map[string]interface{}{
			"category": category,
			"range":    "день",
			"type":     "статистика",
		}
		return result

	} else {
		// Якщо період не визначено, показуємо загальну статистику
		fmt.Println("Показано загальну статистику витрат та доходів")
		// Тут можна додавати реальний код для отримання загальної статистики з бази
		result := map[string]interface{}{
			"category": category,
			"range":    "",
			"type":     "статистика",
		}
		return result
	}
}

func wordToNumber(word string) int {
	numberMap := map[string]int{
		"нуль": 0, "один": 1, "два": 2, "три": 3, "чотири": 4, "п’ять": 5,
		"шість": 6, "сім": 7, "вісім": 8, "дев’ять": 9, "десять": 10,
		"одинадцять": 11, "дванадцять": 12, "тринадцять": 13, "чотирнадцять": 14,
		"п’ятнадцять": 15, "шістнадцять": 16, "сімнадцять": 17, "вісімнадцять": 18,
		"дев’ятнадцять": 19, "двадцять": 20, "тридцять": 30, "сорок": 40,
		"п’ятдесят": 50, "шістдесят": 60, "сімдесят": 70, "вісімдесят": 80,
		"дев’яносто": 90, "сто": 100, "двісті": 200, "триста": 300,
		"чотириста": 400, "п’ятсот": 500, "шістсот": 600, "сімсот": 700,
		"вісімсот": 800, "дев’ятсот": 900, "тисяча": 1000,
	}

	// Normalize the input to lower case for easier matching
	word = strings.ToLower(word)

	if val, found := numberMap[word]; found {
		return val
	}
	return 0 // If the word isn't found, return 0
}

func convertVoiceCommand(command string) int {
	// Split command into parts
	parts := strings.Fields(command)
	total := 0

	// Convert each part into a number and sum up
	for _, part := range parts {
		total += wordToNumber(part)
	}

	return total
}
