package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// проверяем существует ли env переменная с паролем
		pas, ok := os.LookupEnv("TODO_PASSWORD")
		if !ok || pas == "" {
			// если пароль не задан, без проверки
			next(res, req)
			return
		}

		// чекаем куки
		cookie, err := req.Cookie("token")
		// если не найдено, логинимся
		if err != nil {
			writeJsonError(res, fmt.Errorf("Требуется аутентификация"), http.StatusUnauthorized)
			return
		}
		jwtToken := cookie.Value

		// секрет и ожидаемый саб получаем из актуального env todo_password
		secret := sha256.Sum256([]byte(pas))
		expectedSub := hex.EncodeToString(secret[:])

		// Парсим и валидируем токен
		keyFunc := func(t *jwt.Token) (interface{}, error) {
			return secret[:], nil
		}

		var claims jwt.RegisteredClaims
		token, err := jwt.ParseWithClaims(jwtToken, &claims, keyFunc)
		if err != nil || !token.Valid {
			writeJsonError(res, fmt.Errorf("Некорректный токен"), http.StatusUnauthorized)
			return
		}

		if claims.Subject != expectedSub {
			writeJsonError(res, fmt.Errorf("Требуется аутентификация"), http.StatusUnauthorized)
			return
		}
		next(res, req)
	})
}

func signInHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeJsonError(res, fmt.Errorf("Некорректный метод, требуется POST"), http.StatusMethodNotAllowed)
		return
	}

	type signInRequest struct {
		Password string `json:"password"`
	}

	var payload signInRequest

	decode := json.NewDecoder(req.Body)
	if err := decode.Decode(&payload); err != nil {
		writeJsonError(res, fmt.Errorf("Ошибка декодирования данных"), http.StatusBadRequest)
		return
	}

	if payload.Password == "" {
		writeJsonError(res, fmt.Errorf("Поле password не может быть пустым"), http.StatusBadRequest)
		return
	}

	envPas, ok := os.LookupEnv("TODO_PASSWORD")
	if !ok {
		writeJsonError(res, fmt.Errorf("Отсутствует переменная окружения TODO_PASSWORD"), http.StatusBadRequest)
		return
	}

	if envPas != payload.Password {
		writeJsonError(res, fmt.Errorf("Неверный пароль"), http.StatusUnauthorized)
		return
	}

	secret := sha256.Sum256([]byte(payload.Password))
	sub := hex.EncodeToString(secret[:])

	claims := jwt.RegisteredClaims{
		Subject:   sub,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(secret[:])
	if err != nil {
		writeJsonError(res, fmt.Errorf("Не удалось создать токен"), http.StatusUnauthorized)
		return
	}
	writeJson(res, map[string]string{"token": signedToken})
}
