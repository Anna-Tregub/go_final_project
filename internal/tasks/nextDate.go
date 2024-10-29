package tasks

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {

	if repeat == "" {
		return "", fmt.Errorf("не указан repeat")
	}

	nextDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}

	repeatRule := strings.Split(repeat, " ")

	switch repeatRule[0] {

	case "d":

		if len(repeatRule) < 2 {
			return "", fmt.Errorf("не указано количество дней")
		}

		days, err := strconv.Atoi(repeatRule[1])
		if err != nil {
			return "", err
		}

		if days > 400 {
			return "", fmt.Errorf("количество дней не должно превышать 400")
		}
		nextDate := nextDate.AddDate(0, 0, days)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}
		return nextDate.Format("20060102"), nil

	case "y":
		nextDate = nextDate.AddDate(1, 0, 0)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format("20060102"), nil
	default:
		return "", fmt.Errorf("неподдерживаемый формат")
	}

}
