package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/theerudito/stimulsoft_golang/pkg/stimulsoft"
)

var doc = "RPT_Comanda"

func main() {
	mrtPath := fmt.Sprintf("../document/%s.mrt", doc)

	if len(os.Args) > 1 {
		mrtPath = os.Args[1]
	}

	mrt, err := os.ReadFile(mrtPath)
	if err != nil {
		panic(err)
	}

	/*
		jsonBytes := fmt.Sprintf("../document/%s.json", doc)
		if len(os.Args) > 2 {
			jsonBytes = os.Args[2]
		}

		data, err := os.ReadFile(jsonBytes)
		if err != nil {
			panic(err)
		}*/

	jsonBytes, err := json.MarshalIndent(pedido, "", "  ")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	variables := map[string]any{"ExampleVariable": "example"}

	pdf, err := stimulsoft.ExportPDF(ctx, mrt, jsonBytes, "Data", variables)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("report.pdf", pdf, 0644); err != nil {
		panic(err)
	}

	excel, err := stimulsoft.ExportExcel(ctx, mrt, jsonBytes, "Data")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("report.xlsx", excel, 0644); err != nil {
		panic(err)
	}
}
