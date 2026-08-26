package event

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
)

type CertificateData struct {
	StudentName string
	EventTitle  string
	Description string
	Category    string
	Place       string
	Date        string
	EndDate     string
	CheckedInAt string
	Code        string
}

func GenerateCertificatePDF(
	data CertificateData,
) ([]byte, error) {

	pdf := fpdf.New("P", "mm", "A4", "")

	pdf.SetTitle(
		"Certificado de Participação",
		false,
	)

	pdf.AddPage()

	// Borda
	pdf.SetDrawColor(40, 40, 40)
	pdf.SetLineWidth(0.5)

	pdf.Rect(
		15,
		15,
		180,
		267,
		"D",
	)

	// Título
	pdf.SetY(40)

	pdf.SetFont(
		"Arial",
		"B",
		24,
	)

	pdf.CellFormat(
		180,
		15,
		"CERTIFICADO",
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.Ln(10)

	// Texto inicial
	pdf.SetFont(
		"Arial",
		"",
		14,
	)

	pdf.CellFormat(
		180,
		10,
		"Certificamos que",
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.Ln(5)

	// Nome
	pdf.SetFont(
		"Arial",
		"B",
		20,
	)

	pdf.CellFormat(
		180,
		12,
		data.StudentName,
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.Ln(10)

	// Descrição
	pdf.SetFont(
		"Arial",
		"",
		13,
	)

	text := fmt.Sprintf(
		"participou presencialmente do evento \"%s\", "+
			"realizado em %s, no local %s.",
		data.EventTitle,
		data.Date,
		data.Place,
	)

	pdf.MultiCell(
		160,
		8,
		text,
		"",
		"C",
		false,
	)

	pdf.Ln(10)

	// Informações
	pdf.SetFont(
		"Arial",
		"",
		11,
	)

	pdf.CellFormat(
		180,
		8,
		"Categoria: "+data.Category,
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.CellFormat(
		180,
		8,
		"Data da presenca: "+data.CheckedInAt,
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.Ln(20)

	// Código
	pdf.SetFont(
		"Arial",
		"",
		10,
	)

	pdf.CellFormat(
		180,
		7,
		"Codigo de autenticidade:",
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	pdf.SetFont(
		"Arial",
		"B",
		11,
	)

	pdf.CellFormat(
		180,
		8,
		data.Code,
		"",
		1,
		"C",
		false,
		0,
		"",
	)

	var buffer bytes.Buffer

	if err := pdf.Output(&buffer); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
