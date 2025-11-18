package pdf

import (
	"bytes"
	"fmt"

	"test/internal/models"

	"github.com/jung-kurt/gofpdf"
)

func GeneratePDF(ids []uint, rows []models.Link) ([]byte, error) {
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	p.SetFont("Arial", "B", 16)
	p.Cell(0, 10, "Links report")
	p.Ln(12)

	p.SetFont("Arial", "", 11)

	for _, sid := range ids {
		p.SetFont("Arial", "B", 12)
		p.Cell(0, 7, fmt.Sprintf("Set ID: %d", sid))
		p.Ln(7)
		p.SetFont("Arial", "", 10)
		found := false
		for _, r := range rows {
			if r.LinkSetID != sid {
				continue
			}
			found = true
			line := fmt.Sprintf("%s : %s", r.URL, r.Status)
			p.MultiCell(0, 6, line, "", "L", false)
		}
		if !found {
			p.Cell(0, 6, "No links found for this set")
			p.Ln(6)
		}
		p.Ln(4)
	}

	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
