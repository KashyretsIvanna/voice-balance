package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb" // Updated import for speech types
	"cloud.google.com/go/vertexai/genai"
	"github.com/KashyretsIvanna/voice-balance/config"
	"google.golang.org/api/option"
)

type Candidate struct {
	Index            int              `json:"Index"`
	Content          CandidateContent `json:"Content"`
	FinishReason     int              `json:"FinishReason"`
	SafetyRatings    []SafetyRating   `json:"SafetyRatings"`
	FinishMessage    string           `json:"FinishMessage"`
	CitationMetadata interface{}      `json:"CitationMetadata"`
}

type CandidateContent struct {
	Role  string   `json:"Role"`
	Parts []string `json:"Parts"`
}

type SafetyRating struct {
	Category         int     `json:"Category"`
	Probability      int     `json:"Probability"`
	ProbabilityScore float64 `json:"ProbabilityScore"`
	Severity         int     `json:"Severity"`
	SeverityScore    float64 `json:"SeverityScore"`
	Blocked          bool    `json:"Blocked"`
}

type Data struct {
	Candidates     []Candidate `json:"Candidates"`
	PromptFeedback interface{} `json:"PromptFeedback"`
	UsageMetadata  Usage       `json:"UsageMetadata"`
}

type Usage struct {
	PromptTokenCount     int `json:"PromptTokenCount"`
	CandidatesTokenCount int `json:"CandidatesTokenCount"`
	TotalTokenCount      int `json:"TotalTokenCount"`
}

func parseJSONParts(parts []string) []interface{} {
	parsedParts := make([]interface{}, len(parts))
	re := regexp.MustCompile("(?s)```json\\n(.*?)\\n```")
	for i, part := range parts {
		match := re.FindStringSubmatch(part)
		if match != nil {
			var jsonData interface{}
			err := json.Unmarshal([]byte(match[1]), &jsonData)
			if err == nil {
				parsedParts[i] = jsonData
				continue
			}
		}
		parsedParts[i] = part // Leave as-is if not JSON or on error
	}
	return parsedParts
}

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

func AskAi(command string) (error, map[string]interface{}) {
	location := "us-central1"
	modelName := "gemini-1.5-flash-001"
	projectID := "cool-academy-359612"
	ctx := context.Background()

	// Initialize the client with credentials
	client, err := genai.NewClient(ctx, projectID, location, option.WithCredentialsFile(config.Config("CLOUD_JSON_PATH")))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err), nil
	}

	gemini := client.GenerativeModel(modelName)

	// Define the prompt
	prompt := genai.Text(fmt.Sprintf(
		`Я створюю додаток ведення балансу. Ти експерт розпізнавання команд від користувача. 
		Тобі потрібно розпізнати команду та вивести розультат в форматі JSON. Якщо якусь з інформації користувач не надав поверни відповідний ключ з пустою строкою.
		Є кілька типів команд, які підтримує додаток: додавання витрат або 
		доходів, створення нагадувань, статистика. Приклад відовіді яку я 
		очікую, якщо запит на додавання витрат або додавання доходів: 
		{ "amount": 0, "category": "не вказано", "type": "витрати" }. 
		
		Type повинен бути: "доходи", "витрати" або "". 
		Amount: число заокруглене до сотих, category - вказує 
		на що вирати чи доходи(наприклад, продукти).
		Наступний тип команди - створення нагадувань. Приклад 
		відповіді яку я очікую: { "category": "оплатити рахунок за електроенергію", "type": "нагадування" }, 
		де category - текст нагадування. Type - завжди "нагадування". Наступний тип команди. - відобразити статистику. 
		Приклад відповіді яку я очікую: { "category": "", "range": "тиждень", "type": "статистика" }. Повинна повертати range: "тиждень", "рік","місяць",”день”. Якщо не визначено тип команди чи користувач говорить дивні запити, повертай type пустим рядком.

		Розпізнай наступний текст та поверни результат: %s.
		`, command))

	// Generate content
	resp, err := gemini.GenerateContent(ctx, prompt)
	if err != nil {
		return fmt.Errorf("error generating content: %w", err), nil
	}

	// Format the response to JSON
	rb, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("error formatting response to JSON: %w", err), nil
	}

	res := string(rb)
	fmt.Println("Generated Response:", res)

	// Parse the response into a struct
	var data Data
	if err := json.Unmarshal([]byte(res), &data); err != nil {
		return fmt.Errorf("error unmarshalling response into Data struct: %w", err), nil
	}

	// Validate if the response has Candidates and Content
	if len(data.Candidates) == 0 {
		return fmt.Errorf("no candidates found in the response"), nil
	}

	content := &data.Candidates[0].Content
	content.Parts = toStringSlice(parseJSONParts(content.Parts))

	// Ensure at least one Part is available
	if len(content.Parts) == 0 {
		return fmt.Errorf("no content parts found in the response"), nil
	}

	// Parse the JSON from the first part
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(content.Parts[0]), &parsed); err != nil {
		return fmt.Errorf("error unmarshalling content part to JSON: %w", err), nil
	}

	// Return the parsed JSON object
	return nil, parsed
}

func toStringSlice(parts []interface{}) []string {
	result := make([]string, len(parts))
	for i, part := range parts {
		switch v := part.(type) {
		case string:
			result[i] = v
		default:
			jsonPart, _ := json.Marshal(v)
			result[i] = string(jsonPart)
		}
	}
	return result
}
