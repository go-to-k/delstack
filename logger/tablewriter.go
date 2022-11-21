package logger

import (
	"strings"

	"github.com/olekukonko/tablewriter"
)

func ToStringAsTableFormat(header []string, data [][]string) *string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader(header)
	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.AppendBulk(data)
	table.Render()

	stringAsTableFormat := tableString.String()
	return &stringAsTableFormat
}
