package model


// ErrorResponse represents the structure of the error response
type ErrorResponse struct {
	Error string `json:"error"`
}



type TransactionGroupedByCategory struct {
	CategoryName string         `json:"category_name"`
	CategoryType string         `json:"category_type"`
	Transactions []Transaction  `json:"transactions"`
}

