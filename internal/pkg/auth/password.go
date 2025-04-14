package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// MaxPasswordLength - максимальная длина пароля. Она обозначена bcrypt.
// Он не захочет хэшировать пароль, длина байтов которого более 72.
const MaxPasswordLength = 72

// HashPassword - хэширует обычный пароль с помощью bcrypt.
// Возвращает хэшированный пароль в виде строки и ошибку, если она произошла.
// Использует стандартную стоимость хэширования bcrypt (bcrypt.DefaultCost (10)).
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// ComparePassword - сравнивает хэшированный пароль с обычным паролем.
// Возвращает true, если пароли совпадают, в противном случае false.
// Возвращает false так же если хэшированный пароль имеет неверную длину.
func ComparePassword(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
