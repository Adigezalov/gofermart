package utils

import (
	"strconv"
	"strings"
)

// IsValidLuhn проверяет номер по алгоритму Луна
func IsValidLuhn(number string) bool {
	// Убираем пробелы и проверяем, что строка не пустая
	number = strings.ReplaceAll(number, " ", "")
	if len(number) == 0 {
		return false
	}

	// Проверяем, что все символы - цифры
	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Алгоритм Луна требует минимум 2 цифры
	if len(number) < 2 {
		return false
	}

	return luhnChecksum(number) == 0
}

// luhnChecksum вычисляет контрольную сумму по алгоритму Луна
func luhnChecksum(number string) int {
	var sum int
	isEven := false

	// Проходим по цифрам справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))

		if isEven {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum % 10
}

// GenerateValidLuhn генерирует валидный номер по алгоритму Луна
// Добавляет контрольную цифру к переданному префиксу
func GenerateValidLuhn(prefix string) string {
	// Убираем пробелы
	prefix = strings.ReplaceAll(prefix, " ", "")

	// Проверяем, что префикс содержит только цифры
	for _, char := range prefix {
		if char < '0' || char > '9' {
			return "" // Невалидный префикс
		}
	}

	// Вычисляем контрольную цифру
	checkDigit := calculateLuhnCheckDigit(prefix)
	return prefix + strconv.Itoa(checkDigit)
}

// calculateLuhnCheckDigit вычисляет контрольную цифру для префикса
func calculateLuhnCheckDigit(prefix string) int {
	// Добавляем 0 в конец для вычисления
	temp := prefix + "0"
	checksum := luhnChecksum(temp)

	if checksum == 0 {
		return 0
	}
	return 10 - checksum
}
