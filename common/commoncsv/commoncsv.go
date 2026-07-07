package commoncsv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
)

func RecordsToCsv(records []map[string]any) ([]byte, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("there is no record")
	}

	keys := []string{}
	for k := range records[0] {
		keys = append(keys, k)
	}

	var buffer bytes.Buffer
	w := csv.NewWriter(&buffer)
	err := w.Write(keys)
	if err != nil {
		return nil, fmt.Errorf("unable to write keys: %w", err)
	}

	row := make([]string, len(keys))
	for _, r := range records {
		for ki, k := range keys {
			row[ki] = stringify(r[k])
		}

		err := w.Write(row)
		if err != nil {
			return nil, fmt.Errorf("unable to write row: %w", err)
		}
	}

	w.Flush()
	err = w.Error()
	if err != nil {
		return nil, fmt.Errorf("unable to flush csv: %w", err)
	}

	return buffer.Bytes(), nil
}

func stringify(val any) string {
	switch typedVal := val.(type) {
	case float64:
		return strconv.FormatFloat(typedVal, 'f', -1, 64)
	case string:
		return typedVal
	}

	return ""
}
