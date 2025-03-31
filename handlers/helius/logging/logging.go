package logging

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/colintle/crypto-sniper-bot-go/models"
)

var (
	tradeChan chan models.Trade
	once      sync.Once
	stopChan  = make(chan struct{})
	wg        sync.WaitGroup
)

func StartCSVLogger(bufferSize int) {
	once.Do(func() {
		tradeChan = make(chan models.Trade, bufferSize)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case trade := <-tradeChan:
					writeTradeToCSV(trade)
				case <-stopChan:
					return
				}
			}
		}()
	})
}

func LogTradeToCSV(trade models.Trade) {
	if tradeChan == nil {
		log.Println("Trade logger not started. Call StartCSVLogger first.")
		return
	}
	tradeChan <- trade
}

func StopCSVLogger() {
	close(stopChan)
	wg.Wait()
}

func writeTradeToCSV(trade models.Trade) {
	date := trade.Timestamp.Format("2006-01-02")
	filename := filepath.Join("logs", fmt.Sprintf("%s.csv", date))

	_ = os.MkdirAll("logs", 0755)

	_, err := os.Stat(filename)
	newFile := os.IsNotExist(err)

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if newFile {
		writer.Write([]string{
			"Timestamp", "TokenAddress", "Side", "Success", "AmountSOL", "AmountToken", "Balance", "Message", "Error",
		})
	}

	getFloat := func(f *float64) string {
		if f == nil {
			return ""
		}
		return fmt.Sprintf("%.6f", *f)
	}
	getStr := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	record := []string{
		trade.Timestamp.Format(time.RFC3339),
		trade.TokenAddress,
		trade.Side,
		fmt.Sprintf("%t", trade.Success),
		getFloat(trade.AmountSOL),
		getFloat(trade.AmountToken),
		getFloat(trade.Balance),
		getStr(trade.Message),
		getStr(trade.Error),
	}

	if err := writer.Write(record); err != nil {
		log.Printf("Failed to write trade to CSV: %v", err)
	}
}
