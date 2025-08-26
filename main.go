package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

type Expense struct {
	ID          int
	Date        time.Time
	Description string
	Amount      float64
}

type ExpenseTracker struct {
	expenses []Expense
	nextID   int
}

func NewExpenseTracker() *ExpenseTracker {
	return &ExpenseTracker{nextID: 1}
}

func (et *ExpenseTracker) Add(description string, amount float64) int {
	expense := Expense{
		ID:          et.nextID,
		Date:        time.Now(),
		Description: description,
		Amount:      amount,
	}
	et.expenses = append(et.expenses, expense)
	et.nextID++
	return expense.ID
}

func (et *ExpenseTracker) List() {
	if len(et.expenses) == 0 {
		fmt.Println("No expenses yet")
		return
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tDate\tDescription\tAmount")
	for _, e := range et.expenses {
		fmt.Fprintf(tw, "%d\t%s\t%s\t$%.2f\n",
			e.ID, e.Date.Format("2006-01-02"), e.Description, e.Amount)
	}
	_ = tw.Flush()
}

func (et *ExpenseTracker) Summary(month int) {
	total := 0.0
	for _, e := range et.expenses {
		if month == 0 || int(e.Date.Month()) == month {
			total += e.Amount
		}
	}
	if month == 0 {
		fmt.Printf("Total expenses: $%.2f\n", total)
	} else {
		fmt.Printf("Total expenses for %s: $%.2f\n", time.Month(month), total)
	}
}

func (et *ExpenseTracker) Delete(id int) bool {
	for i, e := range et.expenses {
		if e.ID == id {
			et.expenses = append(et.expenses[:i], et.expenses[i+1:]...)
			return true
		}
	}
	return false
}

func main() {
	tracker := NewExpenseTracker()

	// Без аргументов — запустим интерактивный режим (удобно для in-memory)
	if len(os.Args) < 2 {
		fmt.Println("Expense Tracker (in-memory). Type 'help' for commands, 'exit' to quit.")
		repl(tracker)
		return
	}
	runCommand(tracker, os.Args[1:])
}

func runCommand(tracker *ExpenseTracker, args []string) {
	if len(args) == 0 {
		printUsage()
		return
	}

	switch args[0] {
	case "add":
		addCmd := flag.NewFlagSet("add", flag.ContinueOnError)
		desc := addCmd.String("description", "", "Expense description")
		amount := addCmd.Float64("amount", 0, "Expense amount")
		if err := addCmd.Parse(args[1:]); err != nil {
			return
		}
		if *desc == "" || *amount <= 0 {
			fmt.Println("Usage: add --description <text> --amount <number>")
			return
		}
		id := tracker.Add(*desc, *amount)
		fmt.Printf("Expense added successfully (ID: %d)\n", id)

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ContinueOnError)
		_ = listCmd.Parse(args[1:])
		tracker.List()

	case "summary":
		sumCmd := flag.NewFlagSet("summary", flag.ContinueOnError)
		month := sumCmd.Int("month", 0, "Month number (1-12)")
		if err := sumCmd.Parse(args[1:]); err != nil {
			return
		}
		if *month < 0 || *month > 12 {
			fmt.Println("Usage: summary [--month 1..12]")
			return
		}
		tracker.Summary(*month)

	case "delete":
		delCmd := flag.NewFlagSet("delete", flag.ContinueOnError)
		id := delCmd.Int("id", 0, "Expense ID")
		if err := delCmd.Parse(args[1:]); err != nil {
			return
		}
		if *id <= 0 {
			fmt.Println("Usage: delete --id <number>")
			return
		}
		if tracker.Delete(*id) {
			fmt.Println("Expense deleted successfully")
		} else {
			fmt.Println("Expense not found")
		}

	case "help":
		printUsage()
	default:
		fmt.Println("Unknown command:", args[0])
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: expense-tracker <command> [--flags]")
	fmt.Println("Commands:")
	fmt.Println("  add --description <text> --amount <number>")
	fmt.Println("  list")
	fmt.Println("  summary [--month 1..12]")
	fmt.Println("  delete --id <number>")
	fmt.Println("Tip: run without args to enter interactive mode.")
}

func repl(tracker *ExpenseTracker) {
	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !sc.Scan() {
			break
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}
		if line == "help" {
			printUsage()
			continue
		}
		argv := splitArgs(line)
		runCommand(tracker, argv)
	}
}

func splitArgs(s string) []string {
	var args []string
	var b strings.Builder
	inQuotes := false
	escape := false

	for _, r := range s {
		switch {
		case escape:
			b.WriteRune(r)
			escape = false
		case r == '\\':
			escape = true
		case r == '"':
			inQuotes = !inQuotes
		case r == ' ' || r == '\t':
			if inQuotes {
				b.WriteRune(r)
			} else if b.Len() > 0 {
				args = append(args, b.String())
				b.Reset()
			}
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		args = append(args, b.String())
	}
	return args
}
