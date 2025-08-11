#!/bin/bash

# Тест для получения списка заказов пользователя
# GET /api/user/orders

echo "=== Тест получения списка заказов ==="

# Регистрируем пользователя
echo "1. Регистрируем пользователя..."
REGISTER_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"login": "testuser2", "password": "testpass"}')

REGISTER_BODY=$(echo "$REGISTER_RESPONSE" | sed -E 's/HTTPSTATUS:[0-9]{3}$//')
REGISTER_CODE=$(echo "$REGISTER_RESPONSE" | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)

echo "Статус регистрации: $REGISTER_CODE"

if [ "$REGISTER_CODE" != "200" ]; then
    echo "Ошибка регистрации. Пробуем войти..."
    
    # Пробуем войти
    LOGIN_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST http://localhost:8080/api/user/login \
      -H "Content-Type: application/json" \
      -d '{"login": "testuser2", "password": "testpass"}')
    
    LOGIN_BODY=$(echo "$LOGIN_RESPONSE" | sed -E 's/HTTPSTATUS:[0-9]{3}$//')
    LOGIN_CODE=$(echo "$LOGIN_RESPONSE" | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
    
    echo "Статус входа: $LOGIN_CODE"
    
    if [ "$LOGIN_CODE" != "200" ]; then
        echo "Не удалось войти в систему"
        exit 1
    fi
    
    TOKEN=$(echo "$LOGIN_BODY" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
else
    TOKEN=$(echo "$REGISTER_BODY" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
fi

echo "Токен получен: ${TOKEN:0:20}..."

# Проверяем список заказов (должен быть пустой)
echo ""
echo "2. Проверяем пустой список заказов..."
EMPTY_ORDERS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer $TOKEN")

EMPTY_ORDERS_BODY=$(echo "$EMPTY_ORDERS_RESPONSE" | sed -E 's/HTTPSTATUS:[0-9]{3}$//')
EMPTY_ORDERS_CODE=$(echo "$EMPTY_ORDERS_RESPONSE" | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)

echo "Статус: $EMPTY_ORDERS_CODE"
echo "Ответ: $EMPTY_ORDERS_BODY"

if [ "$EMPTY_ORDERS_CODE" = "204" ]; then
    echo "✓ Пустой список заказов возвращает 204 No Content"
else
    echo "✗ Ожидался статус 204, получен $EMPTY_ORDERS_CODE"
fi

# Добавляем заказ
echo ""
echo "3. Добавляем заказ..."
ORDER_NUMBER="79927398713"  # Валидный номер по алгоритму Луна

ADD_ORDER_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: text/plain" \
  -d "$ORDER_NUMBER")

ADD_ORDER_BODY=$(echo "$ADD_ORDER_RESPONSE" | sed -E 's/HTTPSTATUS:[0-9]{3}$//')
ADD_ORDER_CODE=$(echo "$ADD_ORDER_RESPONSE" | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)

echo "Статус добавления заказа: $ADD_ORDER_CODE"
echo "Ответ: $ADD_ORDER_BODY"

if [ "$ADD_ORDER_CODE" = "202" ]; then
    echo "✓ Заказ успешно добавлен"
else
    echo "✗ Ошибка добавления заказа"
fi

# Проверяем список заказов (должен содержать один заказ)
echo ""
echo "4. Проверяем список заказов после добавления..."
ORDERS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer $TOKEN")

ORDERS_BODY=$(echo "$ORDERS_RESPONSE" | sed -E 's/HTTPSTATUS:[0-9]{3}$//')
ORDERS_CODE=$(echo "$ORDERS_RESPONSE" | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)

echo "Статус: $ORDERS_CODE"
echo "Ответ: $ORDERS_BODY"

if [ "$ORDERS_CODE" = "200" ]; then
    echo "✓ Список заказов получен успешно"
    
    # Проверяем, что заказ присутствует в списке
    if echo "$ORDERS_BODY" | grep -q "$ORDER_NUMBER"; then
        echo "✓ Заказ $ORDER_NUMBER найден в списке"
    else
        echo "✗ Заказ $ORDER_NUMBER не найден в списке"
    fi
    
    # Проверяем формат JSON
    if echo "$ORDERS_BODY" | python3 -m json.tool > /dev/null 2>&1; then
        echo "✓ Ответ в корректном JSON формате"
        echo "Форматированный JSON:"
        echo "$ORDERS_BODY" | python3 -m json.tool
    else
        echo "✗ Ответ не в JSON формате"
    fi
else
    echo "✗ Ошибка получения списка заказов"
fi

# Добавляем еще один заказ
echo ""
echo "5. Добавляем второй заказ..."
ORDER_NUMBER2="49927398716"  # Другой валидный номер по алгоритму Луна

ADD_ORDER2_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: text/plain" \
  -d "$ORDER_NUMBER2")

ADD_ORDER2_CODE=$(echo "$ADD_ORDER2_RESPONSE" | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
echo "Статус добавления второго заказа: $ADD_ORDER2_CODE"

# Проверяем список заказов (должен содержать два заказа)
echo ""
echo "6. Проверяем список заказов с двумя заказами..."
FINAL_ORDERS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer $TOKEN")

FINAL_ORDERS_BODY=$(echo "$FINAL_ORDERS_RESPONSE" | sed -E 's/HTTPSTATUS:[0-9]{3}$//')
FINAL_ORDERS_CODE=$(echo "$FINAL_ORDERS_RESPONSE" | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)

echo "Статус: $FINAL_ORDERS_CODE"
echo "Ответ: $FINAL_ORDERS_BODY"

if [ "$FINAL_ORDERS_CODE" = "200" ]; then
    echo "✓ Список заказов получен успешно"
    
    # Проверяем, что оба заказа присутствуют
    if echo "$FINAL_ORDERS_BODY" | grep -q "$ORDER_NUMBER" && echo "$FINAL_ORDERS_BODY" | grep -q "$ORDER_NUMBER2"; then
        echo "✓ Оба заказа найдены в списке"
    else
        echo "✗ Не все заказы найдены в списке"
    fi
    
    # Форматированный вывод
    if echo "$FINAL_ORDERS_BODY" | python3 -m json.tool > /dev/null 2>&1; then
        echo "Форматированный JSON:"
        echo "$FINAL_ORDERS_BODY" | python3 -m json.tool
    fi
fi

# Тест без авторизации
echo ""
echo "7. Тест без авторизации..."
NO_AUTH_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X GET http://localhost:8080/api/user/orders)

NO_AUTH_CODE=$(echo "$NO_AUTH_RESPONSE" | grep -o 'HTTPSTATUS:[0-9]*' | cut -d: -f2)
echo "Статус без авторизации: $NO_AUTH_CODE"

if [ "$NO_AUTH_CODE" = "401" ]; then
    echo "✓ Без авторизации возвращается 401"
else
    echo "✗ Ожидался статус 401, получен $NO_AUTH_CODE"
fi

echo ""
echo "=== Тест завершен ==="