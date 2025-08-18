package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Adigezalov/gophermart/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование:")
		fmt.Println("  go run cmd/luhn-generator/main.go validate <номер>  - проверить номер")
		fmt.Println("  go run cmd/luhn-generator/main.go generate <префикс> - сгенерировать валидный номер")
		fmt.Println("")
		fmt.Println("Примеры:")
		fmt.Println("  go run cmd/luhn-generator/main.go validate 79927398713")
		fmt.Println("  go run cmd/luhn-generator/main.go generate 123456789")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "validate":
		if len(os.Args) < 3 {
			fmt.Println("Ошибка: укажите номер для проверки")
			os.Exit(1)
		}
		number := os.Args[2]

		if utils.IsValidLuhn(number) {
			fmt.Printf("✅ Номер %s ВАЛИДЕН по алгоритму Луна\n", number)
		} else {
			fmt.Printf("❌ Номер %s НЕВАЛИДЕН по алгоритму Луна\n", number)
		}

	case "generate":
		if len(os.Args) < 3 {
			fmt.Println("Ошибка: укажите префикс для генерации")
			os.Exit(1)
		}
		prefix := os.Args[2]

		// Проверяем, что префикс содержит только цифры
		if _, err := strconv.Atoi(prefix); err != nil {
			fmt.Printf("❌ Ошибка: префикс '%s' должен содержать только цифры\n", prefix)
			os.Exit(1)
		}

		validNumber := utils.GenerateValidLuhn(prefix)
		if validNumber == "" {
			fmt.Printf("❌ Ошибка: не удалось сгенерировать номер для префикса '%s'\n", prefix)
			os.Exit(1)
		}

		fmt.Printf("✅ Сгенерирован валидный номер: %s\n", validNumber)
		fmt.Printf("   Префикс: %s\n", prefix)
		fmt.Printf("   Контрольная цифра: %s\n", validNumber[len(prefix):])

	default:
		fmt.Printf("❌ Неизвестная команда: %s\n", command)
		fmt.Println("Доступные команды: validate, generate")
		os.Exit(1)
	}
}
