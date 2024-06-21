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
	Account   Account
	DC        string
	Amount    float64
	Commodity string
}

type Transaction struct {
	// Removed Number from Transaction struct as the slice index will be used.
	Date        string
	Description string
	Entries     []Entry
}

type BalanceSheet struct {
	Party     string
	Asset     map[string]float64
	Liability map[string]float64
	Equity    map[string]float64
}

func NewBalanceSheet() *BalanceSheet {
	return &BalanceSheet{
		Asset:     make(map[string]float64),
		Liability: make(map[string]float64),
		Equity:    make(map[string]float64),
	}
}

type Account struct {
	Party    string
	Category string
	Label    string
}

func updateBalanceSheet(bs *BalanceSheet, entry Entry) {
	switch entry.Account.Category {
	case "Asset":
		switch entry.DC {
		case "D":
			bs.Asset[entry.Commodity] += entry.Amount
		case "C":
			bs.Asset[entry.Commodity] -= entry.Amount
		}
	case "Liability":
		switch entry.DC {
		case "D":
			bs.Liability[entry.Commodity] -= entry.Amount
		case "C":
			bs.Liability[entry.Commodity] += entry.Amount
		}
	case "Equity":
		switch entry.DC {
		case "D":
			bs.Equity[entry.Commodity] -= entry.Amount
		case "C":
			bs.Equity[entry.Commodity] += entry.Amount
		}
	}
}

func parseAccount(accountStr string) (Account, error) {
	parts := strings.Split(accountStr, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return Account{}, fmt.Errorf("invalid account format: %s", accountStr)
	}
	party := parts[0]
	category := parts[1]
	label := strings.Join(parts[1:], ":")
	return Account{Party: party, Category: category, Label: label}, nil
}

type Ledger struct {
	Transactions []Transaction
}

func (l *Ledger) AddTransaction(txn Transaction) {
	l.Transactions = append(l.Transactions, txn)
}

func (l *Ledger) IterateTransactions() []Transaction {
	return l.Transactions
}

func (l *Ledger) BalanceSheetsAt(txnNum int) map[string]*BalanceSheet {
	balanceSheets := make(map[string]*BalanceSheet)
	if txnNum > len(l.Transactions) {
		txnNum = len(l.Transactions) - 1
	}
	for i := 0; i <= txnNum; i++ {
		for _, entry := range l.Transactions[i].Entries {
			party := entry.Account.Party
			bs, exists := balanceSheets[party]
			if !exists {
				bs = NewBalanceSheet()
				bs.Party = party
				balanceSheets[party] = bs
			}
			updateBalanceSheet(bs, entry)
		}
	}
	return balanceSheets
}

func parseLedger(file string) (*Ledger, error) {
	ledger := &Ledger{
		Transactions: []Transaction{},
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var currentDate string
	var currentTransaction Transaction

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		// transaction header starts at column 0
		if len(line) > 0 && line[0] != ' ' {
			if len(currentTransaction.Entries) > 0 {
				ledger.AddTransaction(currentTransaction)
				currentTransaction = Transaction{}
			}
			parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
			currentDate = parts[0]
			currentTransaction.Date = currentDate
			currentTransaction.Description = parts[1]
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}
		accountStr := parts[0]
		dc := parts[1]
		amountStr := parts[2]
		commodity := parts[3]

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing amount: %v", err)
		}

		account, err := parseAccount(accountStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing account: %v", err)
		}

		entry := Entry{
			Date:      currentDate,
			Account:   account,
			DC:        dc,
			Amount:    amount,
			Commodity: commodity,
		}

		currentTransaction.Entries = append(currentTransaction.Entries, entry)
	}

	if len(currentTransaction.Entries) > 0 {
		ledger.AddTransaction(currentTransaction)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return ledger, nil
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
		fmt.Printf("#### %s's Balance Sheet\n", party)

		// print the headings
		head := &Row{Cells: []string{"Asset (Debits)", "Liability (Credits)", "Equity"}}
		fmt.Println(head.Render(columnWidth))
		dashes := strings.Repeat("-", columnWidth)
		div := &Row{Cells: []string{dashes, dashes, dashes}}
		fmt.Println(div.Render(columnWidth))

		assetItems := formatEntries(sheet.Asset)
		liabilityItems := formatEntries(sheet.Liability)
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

		totalAsset := sumMapValues(sheet.Asset)
		totalLiability := sumMapValues(sheet.Liability)
		totalEquity := sumMapValues(sheet.Equity)

		totalAssetStr := fmt.Sprintf("Total: %.2f", totalAsset)
		totalLiabilityStr := fmt.Sprintf("Total: %.2f", totalLiability)
		totalEquityStr := fmt.Sprintf("Total: %.2f", totalEquity)
		totalRow := &Row{Cells: []string{totalAssetStr, totalLiabilityStr, totalEquityStr}}
		fmt.Println(totalRow.Render(columnWidth))

		// ensure the basic accounting equation holds
		if totalAsset != totalLiability+totalEquity {
			fmt.Println("Error: Asset != Liability + Equity")
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
	ledger, err := parseLedger(ledgerFile)
	if err != nil {
		fmt.Println("Failed to parse ledger file:", err)
		return
	}

	for i := range ledger.IterateTransactions() {
		txn := ledger.Transactions[i]
		fmt.Printf("### Transaction %d: %s %s\n\n", i, txn.Date, txn.Description)
		bs := ledger.BalanceSheetsAt(i)
		printBalanceSheet(bs)
	}
}
