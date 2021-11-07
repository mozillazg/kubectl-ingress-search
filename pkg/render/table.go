package render

import (
	"bytes"

	"github.com/mozillazg/kubectl-ingress-search/pkg/ingress"
	"github.com/olekukonko/tablewriter"
)

type TableRender struct {
	NoHeader  bool
	AutoMerge bool
}

func (r *TableRender) Render(rules []ingress.Rule) string {
	buff := bytes.NewBuffer([]byte{})
	table := tablewriter.NewWriter(buff)
	table.SetBorder(false)
	// table.SetCenterSeparator("")
	// table.SetRowSeparator("")
	// table.SetColumnSeparator("")
	// table.SetRowLine(false)
	if r.AutoMerge {
		table.SetAutoMergeCells(true)
		table.SetRowLine(true)
	}
	if !r.NoHeader {
		table.SetHeader([]string{"Namespace", "Name", "Host", "Path", "Backend"})
	}

	for _, r := range rules {
		table.Append([]string{
			r.Namespace.Render(),
			r.Name.Render(),
			r.Host.Render(),
			r.Path.Render(),
			r.Backend.Render(),
		})
	}
	table.Render()
	return buff.String()
}
