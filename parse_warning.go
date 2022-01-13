package jsawn

import "fmt"

type ParseWarning struct {
	Warnings []error
}

func (w ParseWarning) Error() string {
	// pick the first one for now
	warnings := ""
	newline := ""
	for _, warn := range w.Warnings {
		warnings = fmt.Sprintf("%s%s%s", warnings, newline, warn.Error())
		newline = "\n"
	}

	plurality := "warning"
	if len(w.Warnings) > 1 {
		plurality += "s"
	}

	return fmt.Sprintf("%d parse %s\n%s",
		len(w.Warnings),
		plurality,
		warnings,
	)
}
