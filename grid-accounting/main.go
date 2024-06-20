package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Entry struct {
	Date      string
	Account   string
	DC        string
	Amount    float64
	Commodity string
}

type BalanceSheet struct {
	Assets      map[string]float64
	Liabilities map[string]float64
	Equity      map[string]float64
}

func NewBalanceSheet() *BalanceSheet {
	return &BalanceSheet{
		Assets:      make(map[string]float64),
		Liabilities: make(map[string]float64),
		Equity:      make(map[string]float64),
	}
}

func updateBalanceSheet(bs *BalanceSheet, entry Entry) {
	switch entry.DC {
	case "D":
		bs.Assets[entry.Commodity] += entry.Amount
	case "C":
		if strings.HasPrefix(entry.Account, "Equity") {
			bs.Equity[entry.Commodity] += entry.Amount
		} else {
			bs.Liabilities[entry.Commodity] += entry.Amount
		}
	}
}

func parseLedger(file string) map[string]*BalanceSheet {
	entries := make(map[string]*BalanceSheet)

	f, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var currentDate string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 2 {
			currentDate = parts[0]
			continue
		}
		if len(parts) < 4 {
			continue
		}
		commodityAmount := strings.FieldsFunc(parts[3], func(r rune) bool {
			return r == ' ' || r == '.'
		})
		if len(commodityAmount) < 2 {
			continue
		}

		account := strings.Join(parts[0:len(parts)-3], ":")
		dc := parts[len(parts)-3]
		amount, err := strconv.ParseFloat(commodityAmount[0], 64)
		if err != nil {
			fmt.Println("Error parsing amount:", err)
			continue
		}
		commodity := commodityAmount[1]
		party := strings.Split(account, ":")[0]

		entry := Entry{
			Date:      currentDate,
			Account:   account,
			DC:        dc,
			Amount:    amount,
			Commodity: commodity,
		}

		if _, ok := entries[party]; !ok {
			entries[party] = NewBalanceSheet()
		}
		updateBalanceSheet(entries[party], entry)
	}
	return entries
}

func printBalanceSheet(balances map[string]*BalanceSheet) {
	for party, sheet := range balances {
		fmt.Printf("**%s's Balance Sheet**\n", party)
		fmt.Println("| **Assets (Debits)** | **Liabilities (Credits)** | **Equity** |")
		fmt.Println("| --- | --- | --- |")

		for asset, amount := range sheet.Assets {
			fmt.Printf("| %s %.2f | | |\n", asset, amount)
		}
		for liability, amount := range sheet.Liabilities {
			fmt.Printf("| | %s %.2f | |\n", liability, amount)
		}
		for equity, amount := range sheet.Equity {
			fmt.Printf("| | | %s %.2f |\n", equity, amount)
		}

		totalAssets := sumMapValues(sheet.Assets)
		totalLiabilities := sumMapValues(sheet.Liabilities)
		totalEquity := sumMapValues(sheet.Equity)

		fmt.Printf("| **Total**: %.2f | **Total**: %.2f | **Total**: %.2f |\n", totalAssets, totalLiabilities, totalEquity)
		// ensure the basic accounting equation holds
		if totalAssets != totalLiabilities+totalEquity {
			fmt.Println("Error: Assets != Liabilities + Equity")
		}
		fmt.Println()
	}
}

func sumMapValues(m map[string]float64) float64 {
	total := 0.0
	for _, v := range m {
		total += v
	}
	return total
}

func main() {
	ledgerFile := "example.ledger"
	balances := parseLedger(ledgerFile)
	printBalanceSheet(balances)
}
