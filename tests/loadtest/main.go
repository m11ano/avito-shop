package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

var authTokens []string

var handlerStatistics = struct {
	auth     int64
	buy      int64
	sendCoin int64
	info     int64
}{}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	var successCount int64

	/*
		rate := vegeta.Rate{Freq: 1000, Per: time.Second}
		duration := 10 * time.Second
		targeter := vegeta.NewStaticTargeter(vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/api/buy/pen",
			Header: map[string][]string{
				"Authorization": {"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50SUQiOiI2MGFiM2Q2Yy0wMzJiLTQwNzctYmM2MC0yMjAxNjk2OTc0N2MiLCJjcmVhdGVkQXQiOiIxNzM5NTc0MTk4In0.q2GupHDHkR8AjHHOKvBJdiF6ghVVW9Pv4ZwMoZSOkiE"},
			},
		})
		attacker := vegeta.NewAttacker(vegeta.Timeout(10 * time.Second))

		var metrics vegeta.Metrics
		for res := range attacker.Attack(targeter, rate, duration, "Concurrent requests") {
			if res.Code == 200 {
				atomic.AddInt64(&successCount, 1)
			}
			metrics.Add(res)
		}
		metrics.Close()
	*/

	rate := vegeta.Rate{Freq: 1000, Per: time.Second} // 1000 RPS

	// Макс таймаут для запроса
	reqTimeout := 1 * time.Second

	// Длительность этапа добавления пользователей
	stageAccountsTimeDuration := 100 * time.Second

	// Длительность этапа добавления пользователей
	stageRandomActionsTimeDuration := 100 * time.Second

	targeterSignInAndUp := dynamicTargeterSignInAndUp()
	attackerSignInAndUp := vegeta.NewAttacker(vegeta.Timeout(reqTimeout))

	targeterRandomActions := dynamicTargeterRandomActions()
	attackerRandomActions := vegeta.NewAttacker(vegeta.Timeout(reqTimeout))

	var metrics vegeta.Metrics

	authTokens = make([]string, 0)
	mu := sync.Mutex{}

	fmt.Println("Добавляем пользователей...")

	for res := range attackerSignInAndUp.Attack(targeterSignInAndUp, rate, stageAccountsTimeDuration, "Random actions") {
		if res.Code == 200 {
			atomic.AddInt64(&successCount, 1)

			if res.URL == "http://localhost:8080/api/auth" {
				tokenJSON := struct {
					Token string `json:"token"`
				}{}

				//nolint
				_ = json.Unmarshal(res.Body, &tokenJSON)
				if tokenJSON.Token != "" {
					mu.Lock()
					authTokens = append(authTokens, tokenJSON.Token)
					mu.Unlock()
				}
			}
		}

		metrics.Add(res)
	}

	fmt.Println("Перерыв 3 секунды")
	time.Sleep(3 * time.Second)
	fmt.Println("Делаем случайные действия...")

	for res := range attackerRandomActions.Attack(targeterRandomActions, rate, stageRandomActionsTimeDuration, "Random actions") {
		if res.Code == 200 {
			atomic.AddInt64(&successCount, 1)

			if res.URL == "http://localhost:8080/api/auth" {
				tokenJSON := struct {
					Token string `json:"token"`
				}{}

				//nolint
				_ = json.Unmarshal(res.Body, &tokenJSON)
				if tokenJSON.Token != "" {
					mu.Lock()
					authTokens = append(authTokens, tokenJSON.Token)
					mu.Unlock()
				}
			}
		}
		metrics.Add(res)
	}

	metrics.Close()

	// Выводим сводку
	fmt.Printf("Requests: %d, Success: %d (%.2f%%)\n", metrics.Requests, successCount, metrics.Success*100)
	fmt.Printf("Latency [min, mean, p99, max]: %d us, %d us, %d us, %d us\n", metrics.Latencies.Min.Microseconds(), metrics.Latencies.Mean.Microseconds(), metrics.Latencies.P99.Microseconds(), metrics.Latencies.Max.Microseconds())
	fmt.Printf("Status Codes: %v\n", metrics.StatusCodes)

	fmt.Println("Requests statistic:")
	fmt.Printf("auth: %d, buy: %d, sendCoin: %d, info: %d\n", handlerStatistics.auth, handlerStatistics.buy, handlerStatistics.sendCoin, handlerStatistics.info)

	if len(metrics.Errors) > 0 {
		fmt.Printf("Errors: %v\n", metrics.Errors)
	}

	fmt.Println("Done")
}

var usersCount int64

var usersCountInitStage int64

func dynamicTargeterSignInAndUp() vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		c := atomic.AddInt64(&usersCount, 1)
		atomic.AddInt64(&usersCountInitStage, 1)

		body := []byte(fmt.Sprintf(`{"username":"user%d","password":"test"}`, c))

		*tgt = vegeta.Target{
			Method: "POST",
			URL:    "http://localhost:8080/api/auth",
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: body,
		}

		atomic.AddInt64(&handlerStatistics.auth, 1)

		return nil
	}
}

func dynamicTargeterRandomActions() vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		c := atomic.AddInt64(&usersCount, 1)

		variant := rand.IntN(10) + 1

		switch {
		case variant == 1:
			body := []byte(fmt.Sprintf(`{"username":"user%d","password":"test"}`, c))

			*tgt = vegeta.Target{
				Method: "POST",
				URL:    "http://localhost:8080/api/auth",
				Header: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: body,
			}

			atomic.AddInt64(&handlerStatistics.auth, 1)

		case variant > 1 && variant <= 3:
			tokenIdx := rand.IntN(len(authTokens))
			token := authTokens[tokenIdx]

			*tgt = vegeta.Target{
				Method: "GET",
				URL:    "http://localhost:8080/api/buy/pen",
				Header: map[string][]string{
					"Authorization": {token},
				},
			}

			atomic.AddInt64(&handlerStatistics.buy, 1)

		case variant > 3 && variant <= 5:
			tokenIdx := rand.IntN(len(authTokens))
			token := authTokens[tokenIdx]

			body := []byte(fmt.Sprintf(`{"toUser":"user%d","amount":1}`, rand.Int64N(usersCountInitStage)))

			*tgt = vegeta.Target{
				Method: "POST",
				URL:    "http://localhost:8080/api/sendCoin",
				Header: map[string][]string{
					"Content-Type":  {"application/json"},
					"Authorization": {token},
				},
				Body: body,
			}

			atomic.AddInt64(&handlerStatistics.sendCoin, 1)

		default:
			tokenIdx := rand.IntN(len(authTokens))
			token := authTokens[tokenIdx]

			*tgt = vegeta.Target{
				Method: "GET",
				URL:    "http://localhost:8080/api/info",
				Header: map[string][]string{
					"Authorization": {token},
				},
			}

			atomic.AddInt64(&handlerStatistics.info, 1)
		}

		return nil
	}
}
