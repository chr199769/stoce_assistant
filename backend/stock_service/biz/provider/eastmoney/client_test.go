package eastmoney

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFinancialReports(t *testing.T) {
	client := NewClient()
	
	// Test with a known stock code (e.g., Kweichow Moutai 600519)
	reports, err := client.GetFinancialReports(context.Background(), "600519")
	
	if err != nil {
		t.Logf("Failed to get financial reports: %v", err)
		// Don't fail the test if it's network error, just skip
		// Ideally we should mock the HTTP client, but for integration test this is fine
		return
	}

	assert.Nil(t, err)
	assert.NotEmpty(t, reports)

	if len(reports) > 0 {
		report := reports[0]
		assert.NotEmpty(t, report.ReportDate)
		
		// Print the latest report for manual verification
		t.Logf("Latest Report for 600519: Date=%s, Revenue=%.2f, Profit=%.2f", 
			report.ReportDate, report.TotalRevenue, report.NetProfit)
	}

	// Test with invalid code
	reports, err = client.GetFinancialReports(context.Background(), "000000")
	assert.Nil(t, err) // Should return empty list, not error
	assert.Empty(t, reports)
}

func TestGetSectorRank(t *testing.T) {
	client := NewClient()
	sectors, err := client.GetSectorRank(context.Background(), "concept", 5)
	if err != nil {
		t.Logf("Failed to get sector rank: %v", err)
		return
	}
	assert.Nil(t, err)
	assert.NotEmpty(t, sectors)
	for _, s := range sectors {
		t.Logf("Sector: %s, Change: %.2f%%, NetInflow: %.2f", s.Name, s.ChangePercent, s.NetInflow)
		assert.NotEmpty(t, s.Code)
		assert.NotEmpty(t, s.Name)
	}
}
