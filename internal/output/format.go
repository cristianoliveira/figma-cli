package output

import "fmt"

type Format string

const (
	FormatCSV  Format = "csv"
	FormatJSON Format = "json"
	FormatTOON Format = "toon"
)

func ParseFormat(value string, allowed ...Format) (Format, error) {
	format := Format(value)
	for _, candidate := range allowed {
		if format == candidate {
			return format, nil
		}
	}
	return "", fmt.Errorf("invalid --format %q: expected %s", value, formatList(allowed))
}

func formatList(formats []Format) string {
	if len(formats) == 0 {
		return "no formats"
	}
	if len(formats) == 1 {
		return string(formats[0])
	}
	result := string(formats[0])
	for _, format := range formats[1:] {
		result += " or " + string(format)
	}
	return result
}
