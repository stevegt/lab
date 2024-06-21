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

func parseLedger(file string) (map[string]*BalanceSheet, error) {
	entries := make(map[string]*BalanceSheet)

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var currentDate string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "*") {
			parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
			if len(parts) >= 1 {
				currentDate = parts[0]
			}
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}
		account := parts[0]
		dc := parts[1]
		amountStr := parts[2]
		commodity := parts[3]

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing amount: %v", err)
		}
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

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return entries, nil
}

func printBalanceSheet(balances map[string]*BalanceSheet) {
	for party, sheet := range balances {
		fmt.Printf("**%s's Balance Sheet**\n", party)
		fmt.Println("| **Assets (Debits)**               | **Liabilities (Credits)**          | **Equity**                        |")
		fmt.Println("| --------------------------------- | ---------------------------------- | --------------------------------- |")

		maxLenAsset := getMaxLen(sheet.Assets) + 10 // additional space for amount
		maxLenLiability := getMaxLen(sheet.Liabilities) + 10 // additional space for amount
		maxLenEquity := getMaxLen(sheet.Equity) + 10 // additional space for amount
		maxLen := max(maxLenAsset, maxLenLiability, maxLenEquity)

		assetItems := formatEntries(sheet.Assets, maxLen)
		liabilityItems := formatEntries(sheet.Liabilities, maxLen)
		equityItems := formatEntries(sheet.Equity, maxLen)

		maxRows := max(len(assetItems), len(liabilityItems), len(equityItems))

		for i := 0; i < maxRows; i++ {
			var assetStr, liabilityStr, equityStr string

			if i < len(assetItems) {
				assetStr = assetItems[i]
			} else {
				assetStr = fmt.Sprintf("%-*s", maxLen, "")
			}

			if i < len(liabilityItems) {
				liabilityStr = liabilityItems[i]
			} else {
				liabilityStr = fmt.Sprintf("%-*s", maxLen, "")
			}

			if i < len(equityItems) {
				equityStr = equityItems[i]
			} else {
				equityStr = fmt.Sprintf("%-*s", maxLen, "")
			}

			fmt.Printf("| %-*s | %-*s | %-*s |\n", maxLen, assetStr, maxLen, liabilityStr, maxLen, equityStr)
		}

		totalAssets := sumMapValues(sheet.Assets)
		totalLiabilities := sumMapValues(sheet.Liabilities)
		totalEquity := sumMapValues(sheet.Equity)

		fmt.Printf("| **Total**: %-*.2f      | **Total**: %-*.2f      | **Total**: %-*.2f      |\n", maxLen-15, totalAssets, maxLen-15, totalLiabilities, maxLen-15, totalEquity)
		// ensure the basic accounting equation holds
		if totalAssets != totalLiabilities+totalEquity {
			fmt.Println("Error: Assets != Liabilities + Equity")
		}
		fmt.Println()
	}
}

func getMaxLen(m map[string]float64) int {
	maxLen := 0
	for k := range m {
		if len(k) > maxLen {
			maxLen = len(k)
		}
	}
	return maxLen
}

func formatEntries(m map[string]float64, maxLen int) []string {
	items := []string{}
	for commodity, amount := range m {
		items = append(items, fmt.Sprintf("%-*s %.2f", maxLen-10, commodity, amount))
	}
	return items
}

func max(a, b, c int) int {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
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
	balances, err := parseLedger(ledgerFile)
	if err != nil {
		fmt.Println("Failed to parse ledger file:", err)
		return
	}
	printBalanceSheet(balances)
}
