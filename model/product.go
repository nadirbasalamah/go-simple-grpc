package model

// Product struct represents product model
type Product struct {
	ID          int
	Name        string
	Description string
	Category    string
	Amount      int
}

// Products struct represents products model
type Products struct {
	Products []Product
}
