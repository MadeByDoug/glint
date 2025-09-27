// internal/app/infra/output/reporting/reporter.go
package reporting

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// ANSI color codes for terminal output.
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
)

// Formatter prints diagnostics in a specific format.
type Formatter interface {
	Print(w io.Writer, diags []Report)
}

// Reporter orchestrates the printing of diagnostics using a configured formatter.
type Reporter struct {
	formatter Formatter
	min       Severity
}

func New(formatter Formatter, min Severity) *Reporter {
	return &Reporter{formatter: formatter, min: min}
}

func (r *Reporter) SetMinSeverity(threshold Severity) *Reporter {
	r.min = threshold
	return r
}

func (r *Reporter) Emit(w io.Writer, diags []Report) bool {
	if len(diags) == 0 {
		return false
	}

	r.Print(w, diags)
	for _, d := range diags {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

func (r *Reporter) Print(w io.Writer, diags []Report) {
	if w == nil {
		w = os.Stderr
	}
	filtered := r.filterVisible(diags)
	if len(filtered) == 0 {
		return
	}
	r.formatter.Print(w, filtered)
}

func (r *Reporter) filterVisible(diags []Report) []Report {
	if len(diags) == 0 {
		return nil
	}
	includeInfo := r.min == SeverityInfo
	filtered := make([]Report, 0, len(diags))
	for _, d := range diags {
		if d.Severity == SeverityInfo && !includeInfo {
			continue
		}
		filtered = append(filtered, d)
	}
	return filtered
}

// --- Text Formatter ---

type TextFormatter struct {
	useColor bool
}

func NewTextFormatter(useColor bool) *TextFormatter { return &TextFormatter{useColor: useColor} }

func (f *TextFormatter) Print(w io.Writer, diags []Report) {
	palette := f.palette()
	for _, d := range diags {
		if err := f.writeDiag(w, d, palette); err != nil {
			log.Printf("reporter: failed to write diagnostic text: %v", err)
			return
		}
	}
}

// --- JSON Formatter ---

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter { return &JSONFormatter{} }

func (f *JSONFormatter) Print(w io.Writer, diags []Report) {
	enc := json.NewEncoder(w)
	for _, d := range diags {
		out := struct {
			Event    string `json:"event"`
			Severity string `json:"severity"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		}{
			Event:    "diagnostic",
			Severity: string(d.Severity),
			Code:     d.Code,
			Message:  d.Msg,
		}
		if err := enc.Encode(out); err != nil {
			log.Printf("reporter: failed to encode diagnostic to JSON: %v", err)
		}
	}
}

type colorPalette struct {
	reset  string
	bold   string
	colors map[Severity]string
}

func (f *TextFormatter) palette() colorPalette {
	if !f.useColor {
		return colorPalette{}
	}

	return colorPalette{
		reset: Reset,
		bold:  Bold,
		colors: map[Severity]string{
			SeverityError: Red,
			SeverityWarn:  Yellow,
			SeverityInfo:  Cyan,
		},
	}
}

func (f *TextFormatter) writeDiag(w io.Writer, d Report, palette colorPalette) error {
	if _, err := fmt.Fprintf(w, "%s%s%s%s", palette.bold, palette.colorFor(d.Severity), d.Severity, palette.reset); err != nil {
		return fmt.Errorf("write diagnostic severity: %w", err)
	}

	if d.Code != "" {
		if _, err := fmt.Fprintf(w, "%s[%s]%s", palette.bold, d.Code, palette.reset); err != nil {
			return fmt.Errorf("write diagnostic code: %w", err)
		}
	}

	if _, err := fmt.Fprintf(w, ": %s\n", d.Msg); err != nil {
		return fmt.Errorf("write diagnostic message: %w", err)
	}
	return nil
}

func (p colorPalette) colorFor(sev Severity) string {
	if p.colors == nil {
		return ""
	}
	if color, ok := p.colors[sev]; ok {
		return color
	}
	return ""
}
