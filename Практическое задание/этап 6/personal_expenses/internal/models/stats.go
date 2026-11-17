package models

// Сумма трат по одной категории
type CategoryStats struct {
		CategoryName string
		Total		 float64
}

// Сумма трат за один месяц вида "2025-03"
type MonthStats struct {
		Month string
		Total float64
}