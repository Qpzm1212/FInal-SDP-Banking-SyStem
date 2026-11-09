package patterns

import "errors"

// ErrInvalidInput - общая ошибка валидации (API превращает её в HTTP 400).
var ErrInvalidInput = errors.New("некорректные данные")
