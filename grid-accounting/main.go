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
	Asset     map[string]map[string]float64
	Liability map[string]map[string]float64
	Equity    map[string]map[string]float64
}

func NewBalanceSheet() *BalanceSheet {
	return &BalanceSheet{
		Asset:     make(map[string]map[string]float64),
		Liability: make(map[string]map[string]float64),
		Equity:    make(map[string]map[string]float64),
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
		if bs.Asset[entry.Commodity] == nil {
			bs.Asset[entry.Commodity] = make(map[string]float64)
		}
		switch entry.DC {
		case "D":
			bs.Asset[entry.Commodity][entry.Account.Label] += entry.Amount
		case "C":
			bs.Asset[entry.Commodity][entry.Account.Label] -= entry.Amount
		}
	case "Liability":
		if bs.Liability[entry.Commodity] == nil {
			bs.Liability[entry.Commodity] = make(map[string]float64)
		}
		switch entry.DC {
		case "D":
			bs.Liability[entry.Commodity][entry.Account.Label] -= entry.Amount
		case "C":
			bs.Liability[entry.Commodity][entry.Account.Label] += entry.Amount
		}
	case "Equity":
		if bs.Equity[entry.Commodity] == nil {
			bs.Equity[entry.Commodity] = make(map[string]float64)
		}
		switch entry.DC {
		case "D":
			bs.Equity[entry.Commodity][entry.Account.Label] -= entry.Amount
		case "C":
			bs.Equity[entry.Commodity][entry.Account.Label] += entry.Amount
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
	label := strings.Join(parts[2:], ":")
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
		lineTrimmed := strings.TrimSpace(line)

		// skip comments
		if strings.HasPrefix(lineTrimmed, "#") {
			continue
		}

		// skip empty lines
		if len(lineTrimmed) == 0 {
			continue
		}

		// transaction header starts at column 0
		if line[0] != ' ' {
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
			return nil, fmt.Errorf("invalid entry format: %s", line)
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
	columnWidth := 30
	for party, sheet := range balances {
		fmt.Printf("#### %s's Balance Sheet\n", party)

		// print the headings
		head := &Row{Cells: []string{"Asset (Debits)", "Liability (Credits)", "Equity"}}
		fmt.Println(head.Render(columnWidth))
		dashes := strings.Repeat("-", columnWidth)
		div := &Row{Cells: []string{dashes, dashes, dashes}}
		fmt.Println(div.Render(columnWidth))

		// collect all the commodities from each category
		commodities := make(map[string]bool)
		for commodity := range sheet.Asset {
			commodities[commodity] = true
		}
		for commodity := range sheet.Liability {
			commodities[commodity] = true
		}
		for commodity := range sheet.Equity {
			commodities[commodity] = true
		}

		// create a totals map to hold the total for each commodity and category
		// map[commodity]map[category]total
		totals := make(map[string]map[string]float64)
		for commodity := range commodities {
			totals[commodity] = make(map[string]float64)
			totals[commodity]["Asset"] = 0
			totals[commodity]["Liability"] = 0
			totals[commodity]["Equity"] = 0
		}

		// print the entries for each commodity
		for commodity := range commodities {
			subhead := &Row{Cells: []string{fmt.Sprintf("%s:", commodity), "", ""}}
			fmt.Println(subhead.Render(columnWidth))

			assetEntries := sheet.Asset[commodity]
			liabilityEntries := sheet.Liability[commodity]
			equityEntries := sheet.Equity[commodity]

			assetItems := formatEntries(assetEntries, commodity)
			liabilityItems := formatEntries(liabilityEntries, commodity)
			equityItems := formatEntries(equityEntries, commodity)

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

			totalAsset := sumMapValues(assetEntries)
			totalLiability := sumMapValues(sheet.Liability[commodity])
			totalEquity := sumMapValues(sheet.Equity[commodity])

			totalAssetStr := fmt.Sprintf("Total: %.2f %s", totalAsset, commodity)
			totalLiabilityStr := fmt.Sprintf("Total: %.2f %s", totalLiability, commodity)
			totalEquityStr := fmt.Sprintf("Total: %.2f %s", totalEquity, commodity)
			totalRow := &Row{Cells: []string{totalAssetStr, totalLiabilityStr, totalEquityStr}}
			fmt.Println(totalRow.Render(columnWidth))

			// ensure the basic accounting equation holds
			if totalAsset != totalLiability+totalEquity {
				fmt.Println("Error: Asset != Liability + Equity")
			}

			// add a blank row
			blanks := &Row{Cells: []string{"", "", ""}}
			fmt.Println(blanks.Render(columnWidth))
		}
		fmt.Println()
	}
}

func formatEntries(m map[string]float64, commodity string) []string {
	items := []string{}
	for accountLabel, amount := range m {
		items = append(items, fmt.Sprintf("%-10s %.2f %s", accountLabel, amount, commodity))
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
