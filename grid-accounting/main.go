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
		fmt.Println("Error opening file:", err)
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
		fmt.Println("Processing line:", line)
		parts := strings.SplitN(strings.TrimSpace(line), " ", 4)
		if len(parts) == 2 {
			currentDate = parts[0]
			fmt.Println("Current date set to:", currentDate)
			continue
		}
		if len(parts) < 4 {
			fmt.Println("Skipping incomplete line:", line)
			continue
		}
		account := parts[0]
		dc := parts[1]
		amountStr := parts[2]
		commodity := parts[3]

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			fmt.Println("Error parsing amount:", err)
			continue
		}
		party := strings.Split(account, ":")[0]

		entry := Entry{
			Date:      currentDate,
			Account:   account,
			DC:        dc,
			Amount:    amount,
			Commodity: commodity,
		}

		fmt.Println("Parsed entry:", entry)

		if _, ok := entries[party]; !ok {
			entries[party] = NewBalanceSheet()
		}
		updateBalanceSheet(entries[party], entry)
		fmt.Println("Updated balance sheet for party:", party)
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
	fmt.Println("Parsing ledger file:", ledgerFile)
	balances := parseLedger(ledgerFile)
	if balances == nil {
		fmt.Println("Failed to parse ledger file.")
		return
	}
	fmt.Println("Printing balance sheet...")
	printBalanceSheet(balances)
}
