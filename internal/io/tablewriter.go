package io

import (
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

func ToStringAsTableFormat(header []string, data [][]string) (*string, error) {
	tableString := &strings.Builder{}
	table := tablewriter.NewTable(tableString,
		tablewriter.WithRendition(
			tw.Rendition{
				Symbols: tw.NewSymbols(tw.StyleASCII),
				Borders: tw.Border{
					Top:    tw.On,
					Bottom: tw.On,
					Left:   tw.On,
					Right:  tw.On,
				},
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenRows: tw.On,
					},
					Lines: tw.Lines{
						ShowHeaderLine: tw.On,
					},
				},
			},
		),
	)

	table.Header(header)
	err := table.Bulk(data)
	if err != nil {
		return nil, err
	}
	err = table.Render()
	if err != nil {
		return nil, err
	}

	stringAsTableFormat := tableString.String()
	return &stringAsTableFormat, nil
}
