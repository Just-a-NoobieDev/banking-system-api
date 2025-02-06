package pdf

import (
	"banking-system/internal/database/models"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-pdf/fpdf"
)

type StatementGenerator struct {
	outputDir    string
	bankName    string
	bankAddress string
	bankContact string
}

// Constructor with configuration
func NewStatementGenerator(config StatementConfig) *StatementGenerator {
	return &StatementGenerator{
		outputDir:    config.OutputDir,
		bankName:    config.BankName,
		bankAddress: config.BankAddress,
		bankContact: config.BankContact,
	}
}

// Use a config struct for flexibility
type StatementConfig struct {
	OutputDir    string
	BankName    string
	BankAddress string
	BankContact string
}

// Add method to customize colors
type Colors struct {
	Primary   [3]int
	Secondary [3]int
	Text      [3]int
}

func (g *StatementGenerator) setColors(pdf *fpdf.Fpdf, colors Colors) {
	pdf.SetTextColor(colors.Text[0], colors.Text[1], colors.Text[2])
}

// Add method to handle errors during PDF generation
func (g *StatementGenerator) handleError(err error, operation string) error {
	if err != nil {
		return fmt.Errorf("PDF generation failed during %s: %w", operation, err)
	}
	return nil
}

func (g *StatementGenerator) GenerateStatement(transactions []models.Transaction, totalAmount float64, userID int, userFullName string) (string, error) {
	pdf := g.initializePDF()
	
	g.addHeader(pdf)
	g.addBankInfo(pdf)
	g.addStatementInfo(pdf, userID, userFullName, transactions[0].AccountID)
	g.addTransactionTable(pdf, transactions)
	g.addSummarySection(pdf, totalAmount)
	g.addFooter(pdf)
	
	return g.saveFile(pdf, userID)
}

func (g *StatementGenerator) initializePDF() *fpdf.Fpdf {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()
	return pdf
}

func (g *StatementGenerator) addHeader(pdf *fpdf.Fpdf) {
	// Add logo if available
	// pdf.Image("path/to/logo.png", 20, 20, 30, 0, false, "", 0, "")
	
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(0, 48, 87) // Navy blue
	pdf.Cell(170, 10, "Statement of Account")
	pdf.Ln(15)
}

func (g *StatementGenerator) addBankInfo(pdf *fpdf.Fpdf) {
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(70, 70, 70)
	pdf.Cell(170, 5, g.bankName)
	pdf.Ln(5)
	pdf.Cell(170, 5, g.bankAddress)
	pdf.Ln(5)
	pdf.Cell(170, 5, g.bankContact)
	pdf.Ln(15)
}

func (g *StatementGenerator) addStatementInfo(pdf *fpdf.Fpdf, userID int, userFullName string, accountID int) {
	pdf.SetFillColor(245, 245, 245)
	pdf.Rect(20, pdf.GetY(), 170, 40, "F")
	
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(70, 70, 70)
	
	x := pdf.GetX()
	y := pdf.GetY() + 5
	
	// Left column
	pdf.SetXY(x+5, y)
	pdf.Cell(40, 5, "Account Holder:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 5, userFullName)
	
	// Right column
	pdf.SetXY(x+95, y)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 5, "Statement Date:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 5, time.Now().Format("January 2, 2006"))
	
	// Second row
	pdf.SetXY(x+5, y+10)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 5, "Account Number:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 5, fmt.Sprintf("%d", accountID))
	
	// Right column
	pdf.SetXY(x+95, y+10)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 5, "Period:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 5, fmt.Sprintf("%s - %s", 
		time.Now().AddDate(0, -1, 0).Format("Jan 2"),
		time.Now().Format("Jan 2, 2006")))
	
	pdf.Ln(15)
}

func (g *StatementGenerator) addTransactionTable(pdf *fpdf.Fpdf, transactions []models.Transaction) {
	pdf.Ln(10)
	pdf.SetFillColor(0, 48, 87)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 10)
	
	headers := []string{"Date", "Reference Id", "Transaction Id", "Type", "Amount"}
	widths := []float64{30, 65, 30, 30, 25}
	
	for i, header := range headers {
		pdf.Cell(widths[i], 8, header)
	}
	pdf.Ln(-1)
	
	pdf.SetTextColor(70, 70, 70)
	pdf.SetFont("Arial", "", 9)
	
	for i, trans := range transactions {
		if i%2 == 0 {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		
		pdf.Cell(widths[0], 7, trans.CreatedAt.Format("02/01/2006"))
		pdf.Cell(widths[1], 7, trans.ReferenceID)
		pdf.Cell(widths[2], 7, strconv.Itoa(trans.ID))
		pdf.Cell(widths[3], 7, string(trans.Type))
		
		amount := fmt.Sprintf("%.2f", trans.Amount)
		if trans.Type == models.Withdrawal {
			pdf.SetTextColor(200, 0, 0)
		} else {
			pdf.SetTextColor(0, 150, 0)
		}
		pdf.Cell(widths[4], 7, amount)
		pdf.SetTextColor(70, 70, 70)
		
		pdf.Ln(-1)
	}
}

func (g *StatementGenerator) addSummarySection(pdf *fpdf.Fpdf, totalAmount float64) {
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(0, 48, 87)
	
	pdf.Cell(140, 8, "Closing Balance:")
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(30, 8, fmt.Sprintf("%.2f", totalAmount))
	pdf.Ln(15)
}

func (g *StatementGenerator) addFooter(pdf *fpdf.Fpdf) {
	pdf.SetY(-40)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(170, 5, "This is a computer generated statement and doesn't require signature.")
	pdf.Ln(5)
	pdf.Cell(170, 5, "For any queries, please contact our customer service.")
}

func (g *StatementGenerator) saveFile(pdf *fpdf.Fpdf, userID int) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("statement_%d_%s.pdf", userID, timestamp)
	
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return "", g.handleError(err, "creating output directory")
	}
	
	filePath := filepath.Join(g.outputDir, fileName)
	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return "", g.handleError(err, "saving PDF")
	}
	
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", g.handleError(err, "getting absolute path")
	}
	
	return absPath, nil
}

func getTransactionDescription(trans models.Transaction) string {
	// You can implement custom logic here to generate meaningful descriptions
	// based on transaction type, reference, or other attributes
	return "Transaction " + trans.ReferenceID
}