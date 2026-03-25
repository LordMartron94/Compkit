package validation

import "syntaxa"

/*
ValidationSeverity indicates how serious a validation entry is.

Use cases:
- Collecting diagnostics during compilation (parser/transform/semantic passes)
- Differentiating between informational notes and errors that should stop a pipeline
*/
type ValidationSeverity uint8

const (
	VALIDATION_DIAGNOSTIC ValidationSeverity = iota + 1
	VALIDATION_INFO
	VALIDATION_NOTICE
	VALIDATION_WARNING
	VALIDATION_ERROR
	VALIDATION_FATAL

	validationSeverityCount
)

/* String returns a fixed-width name for the severity (for logging and display). */
func (v ValidationSeverity) String() string {
	switch v {
	case VALIDATION_DIAGNOSTIC:
		return "DIAGNOSTIC"
	case VALIDATION_INFO:
		return "INFO"
	case VALIDATION_NOTICE:
		return "NOTICE"
	case VALIDATION_WARNING:
		return "WARNING"
	case VALIDATION_ERROR:
		return "ERROR"
	case VALIDATION_FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

/*
ValidationCode is a stable machine-readable identifier for a validation entry.

The intent is that callers can filter, suppress, aggregate, or map these codes to UI/help text.
*/
type ValidationCode string

/*
ValidationEntry is one validation finding.

Location is a Syntaxa span so findings can be mapped back to source text.
*/
type ValidationEntry struct {
	Severity ValidationSeverity

	Code    ValidationCode
	Message string
	Note    *string

	Location syntaxa.Span
}

/*
ValidationEntries stores validation findings and supports severity queries.

This structure is designed as an append-only collector.
*/
type ValidationEntries struct {
	entries []ValidationEntry
	has     []bool
}

/*
ValidationEntriesCreate returns a ready-to-use validation collector.

Time complexity: O(1)
Space complexity: O(1)
*/
func ValidationEntriesCreate() *ValidationEntries {
	return &ValidationEntries{
		entries: make([]ValidationEntry, 0),
		has:     make([]bool, validationSeverityCount),
	}
}

/*
ValidationEntriesGet returns a copy of the collected entries.

Time complexity: O(n)
Space complexity: O(n)
*/
func ValidationEntriesGet(v *ValidationEntries) []ValidationEntry {
	if v == nil {
		return nil
	}

	cp := make([]ValidationEntry, len(v.entries))
	copy(cp, v.entries)
	return cp
}

/*
ValidationEntriesAddEntry appends one entry to the collector.

Prerequisites:
- v must be non-nil
- entry.Severity must be a valid ValidationSeverity
*/
func ValidationEntriesAddEntry(v *ValidationEntries, entry ValidationEntry) {
	if v == nil {
		panic("validation entries is nil -> cannot add entry")
	}

	if entry.Severity <= 0 || entry.Severity >= validationSeverityCount {
		panic("invalid validation severity")
	}

	v.entries = append(v.entries, entry)
	v.has[entry.Severity] = true
}

/*
ValidationEntriesHasExact reports whether there is at least one entry with exactly severity.
*/
func ValidationEntriesHasExact(v *ValidationEntries, severity ValidationSeverity) bool {
	if v == nil {
		return false
	}
	if severity <= 0 || severity >= validationSeverityCount {
		return false
	}
	return v.has[severity]
}

/*
ValidationEntriesHasOrHigher reports whether there is at least one entry with severity or higher.
*/
func ValidationEntriesHasOrHigher(v *ValidationEntries, severity ValidationSeverity) bool {
	if v == nil {
		return false
	}
	if severity <= 0 || severity >= validationSeverityCount {
		return false
	}

	for i := severity; i < validationSeverityCount; i++ {
		if v.has[i] {
			return true
		}
	}

	return false
}

/*
ValidationEntriesAdd creates and adds a new entry with the given severity.
*/
func ValidationEntriesAdd(
	v *ValidationEntries,
	severity ValidationSeverity,
	code ValidationCode,
	message string,
	note *string,
	location syntaxa.Span,
) {
	ValidationEntriesAddEntry(v, ValidationEntry{
		Severity: severity,
		Code:     code,
		Message:  message,
		Note:     note,
		Location: location,
	})
}

/*
ValidationEntriesAddDiagnostic adds a DIAGNOSTIC entry.
*/
func ValidationEntriesAddDiagnostic(v *ValidationEntries, code ValidationCode, message string, note *string, location syntaxa.Span) {
	ValidationEntriesAdd(v, VALIDATION_DIAGNOSTIC, code, message, note, location)
}

/*
ValidationEntriesAddInfo adds an INFO entry.
*/
func ValidationEntriesAddInfo(v *ValidationEntries, code ValidationCode, message string, note *string, location syntaxa.Span) {
	ValidationEntriesAdd(v, VALIDATION_INFO, code, message, note, location)
}

/*
ValidationEntriesAddNotice adds a NOTICE entry.
*/
func ValidationEntriesAddNotice(v *ValidationEntries, code ValidationCode, message string, note *string, location syntaxa.Span) {
	ValidationEntriesAdd(v, VALIDATION_NOTICE, code, message, note, location)
}

/*
ValidationEntriesAddWarning adds a WARNING entry.
*/
func ValidationEntriesAddWarning(v *ValidationEntries, code ValidationCode, message string, note *string, location syntaxa.Span) {
	ValidationEntriesAdd(v, VALIDATION_WARNING, code, message, note, location)
}

/*
ValidationEntriesAddError adds an ERROR entry.
*/
func ValidationEntriesAddError(v *ValidationEntries, code ValidationCode, message string, note *string, location syntaxa.Span) {
	ValidationEntriesAdd(v, VALIDATION_ERROR, code, message, note, location)
}

/*
ValidationEntriesAddFatal adds a FATAL entry.
*/
func ValidationEntriesAddFatal(v *ValidationEntries, code ValidationCode, message string, note *string, location syntaxa.Span) {
	ValidationEntriesAdd(v, VALIDATION_FATAL, code, message, note, location)
}
