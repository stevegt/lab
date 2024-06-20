package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Entry struct {
	Date     string
	Account  string
	DC       string
	Amount   string
	Commodity string
}

type BalanceSheet struct {
	Assets       map[string]string
	Liabilities  map[string]string
	Equity       map[string]string
}

func NewBalanceSheet() *BalanceSheet {
	return &BalanceSheet{
		Assets:       make(map[string]string),
		Liabilities:  make(map[string]string),
		Equity:       make(map[string]string),
	}
}

func updateBalanceSheet(bs *BalanceSheet, entry Entry) {
	switch entry.DC {
	case "D":
		bs.Assets[entry.Commodity] = entry.Amount
	case "C":
		bs.Liabilities[entry.Commodity] = entry.Amount
		if entry.Commodity == strings.Split(entry.Account, ":")[2] {
			bs.Equity[entry.Commodity] = entry.Amount
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

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}
		date := parts[0]
		account := parts[1]
		dc := parts[2]
		amount := parts[3]
		commodity := parts[4]

		entry := Entry{
			Date:     date,
			Account:  account,
			DC:       dc,
			Amount:   amount,
			Commodity: commodity,
		}

		party := strings.Split(account, ":")[0]
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
			fmt.Printf("| %s | | |\n", amount)
		}
		for liability, amount := range sheet.Liabilities {
			fmt.Printf("| | %s | |\n", amount)
		}
		for equity, amount := range sheet.Equity {
			fmt.Printf("| | | %s |\n", amount)
		}
		fmt.Println()
	}
}

func main() {
	ledgerFile := "example.ledger"
	balances := parseLedger(ledgerFile)
	printBalanceSheet(balances)
}
