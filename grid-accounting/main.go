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
		if strings.HasPrefix(entry.Account, "Assets") {
			bs.Assets[entry.Commodity] += entry.Amount
		} else if strings.HasPrefix(entry.Account, "Liabilities") || strings.HasPrefix(entry.Account, "Equity") {
			bs.Liabilities[entry.Commodity] -= entry.Amount // Decrease for liabilities or equity
		}
	case "C":
		if strings.HasPrefix(entry.Account, "Assets") {
			bs.Assets[entry.Commodity] -= entry.Amount
		} else if strings.HasPrefix(entry.Account, "Liabilities") {
			bs.Liabilities[entry.Commodity] += entry.Amount
		} else if strings.HasPrefix(entry.Account, "Equity") {
			bs.Equity[entry.Commodity] += entry.Amount
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

// Row is a row in the balance sheet
type Row struct {
	Cells []string
}

// Render renders the row as a markdown table row
func (r *Row) Render(columnWidth int) string {
	var cells []string
	for _, cell := range r.Cells {
		spaces := columnWidth - len(cell)
		spaces = max(0, spaces)
		cells = append(cells, fmt.Sprintf("%s%s", cell, strings.Repeat(" ", spaces)))
	}
	return fmt.Sprintf("| %s |", strings.Join(cells, " | "))
}

func printBalanceSheet(balances map[string]*BalanceSheet) {
	columnWidth := 21
	for party, sheet := range balances {
		fmt.Printf("**%s's Balance Sheet**\n", party)

		// print the headings
		head := &Row{Cells: []string{"Assets (Debits)", "Liabilities (Credits)", "Equity"}}
		fmt.Println(head.Render(columnWidth))
		dashes := strings.Repeat("-", columnWidth)
		div := &Row{Cells: []string{dashes, dashes, dashes}}
		fmt.Println(div.Render(columnWidth))

		assetItems := formatEntries(sheet.Assets)
		liabilityItems := formatEntries(sheet.Liabilities)
		equityItems := formatEntries(sheet.Equity)

		maxRows := maxItems(len(assetItems), len(liabilityItems), len(equityItems))

		for i := 0; i < maxRows; i++ {
			var assetStr, liabilityStr, equityStr string

			if i < len(assetItems) {
				assetStr = assetItems[i]
			}

			if i < len(liabilityItems) {
				liabilityStr = liabilityItems[i]
			}

			if i < len(equityItems) {
				equityStr = equityItems[i]
			}
			row := &Row{Cells: []string{assetStr, liabilityStr, equityStr}}
			fmt.Println(row.Render(columnWidth))
		}

		totalAssets := sumMapValues(sheet.Assets)
		totalLiabilities := sumMapValues(sheet.Liabilities)
		totalEquity := sumMapValues(sheet.Equity)

		totalAssetsStr := fmt.Sprintf("Total: %.2f", totalAssets)
		totalLiabilitiesStr := fmt.Sprintf("Total: %.2f", totalLiabilities)
		totalEquityStr := fmt.Sprintf("Total: %.2f", totalEquity)
		totalRow := &Row{Cells: []string{totalAssetsStr, totalLiabilitiesStr, totalEquityStr}}
		fmt.Println(totalRow.Render(columnWidth))

		// ensure the basic accounting equation holds
		if totalAssets != totalLiabilities+totalEquity {
			fmt.Println("Error: Assets != Liabilities + Equity")
		}
		fmt.Println()
	}
}

func formatEntries(m map[string]float64) []string {
	items := []string{}
	for commodity, amount := range m {
		items = append(items, fmt.Sprintf("%-10s %.2f", commodity, amount))
	}
	return items
}

func maxItems(a, b, c int) int {
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
