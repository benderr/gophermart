package moonvalidator

import (
	"errors"
	"strconv"
)

func MoonValidator(number string) error {
	arr := []rune(number)
	var resultSum int64 = 0
	parity := len(arr) % 2

	for i, r := range arr {
		val, err := strconv.ParseInt(string(r), 10, 64)

		if err != nil {
			return err
		}

		del := i % 2

		if del == parity {
			m := val * 2
			if m > 9 {
				resultSum += m - 9
			} else {
				resultSum += m
			}
		} else {
			resultSum += val
		}
	}

	if resultSum%10 == 0 {
		return nil
	}

	return ErrInvalidNumber
}

var (
	ErrInvalidNumber = errors.New("invalid number")
)
