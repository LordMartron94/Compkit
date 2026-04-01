package validation

import (
	"fmt"
	"io"
	"strings"
)

type columnAdvanceFn func(rune, int) int

/*
RenderValidationEntriesWithContext writes a formatted report of validation entries with source context.

Use cases:
- Printing compiler diagnostics to a terminal (human readable)
- Emitting a validation report during development or CI

Time complexity: O(n + S) where n is entry count and S is rendered source span size
Space complexity: O(L) where L is number of source lines (split for rendering)
*/
func RenderValidationEntriesWithContext(
	w io.Writer,
	source []rune,
	entries *ValidationEntries,
	advanceFn columnAdvanceFn,
) {
	if w == nil || entries == nil {
		return
	}

	list := ValidationEntriesGet(entries)
	if len(list) == 0 {
		return
	}

	if advanceFn == nil {
		advanceFn = func(r rune, col int) int {
			if r == '\t' {
				return col + 4
			}
			return col + 1
		}
	}

	lines := splitLinesRunes(source)

	fmt.Fprintln(w, "\n===== VALIDATION =====")

	for _, entry := range list {
		renderValidationEntry(w, entry, lines, advanceFn)
	}

	fmt.Fprintln(w, "======================")
}

func renderValidationEntry(w io.Writer, entry ValidationEntry, lines [][]rune, advanceFn columnAdvanceFn) {
	fmt.Fprintf(w, "[%v] %s — %s\n", entry.Severity, entry.Code, entry.Message)
	if entry.Note != nil && *entry.Note != "" {
		fmt.Fprintf(w, "      Note: %s\n", *entry.Note)
	}

	startL := entry.Location.StartLine
	startC := entry.Location.StartColumn
	endL := entry.Location.EndLine
	endC := entry.Location.EndColumn

	if startL <= 0 || endL <= 0 || startC <= 0 || endC <= 0 {
		if entry.Location.Start != 0 || entry.Location.End != 0 {
			fmt.Fprintf(w, "      Absolute Span: %d - %d\n", entry.Location.Start, entry.Location.End)
		} else {
			fmt.Fprintln(w, "      Location: (no source span)")
		}
		fmt.Fprintln(w, " ──────────────────────────────────────────────────────────")
		return
	}

	fmt.Fprintf(w, "      Location: line %d:%d to %d:%d\n", startL, startC, endL, endC)
	renderDiagnosticContext(w, lines, startL, startC, endL, endC, advanceFn)
}

func renderDiagnosticContext(w io.Writer, lines [][]rune, startL, startC, endL, endC int, advanceFn columnAdvanceFn) {
	if !isValidSpanRange(startL, endL, len(lines)) {
		renderInvalidSpanBlock(w, lines, startL, startC, endL, endC, advanceFn)
		return
	}

	startC = enforceMinimumColumn(startC)
	endC = enforceMinimumColumn(endC)

	if startL == endL {
		renderSingleLineHighlight(w, lines[startL-1], startL, startC, endC, advanceFn)
		fmt.Fprintln(w, " ──────────────────────────────────────────────────────────")
		return
	}

	renderMultiLineHighlight(w, lines, startL, startC, endL, endC, advanceFn)
	fmt.Fprintln(w, " ──────────────────────────────────────────────────────────")
}

func isValidSpanRange(startLine, endLine, totalLines int) bool {
	if startLine <= 0 || endLine <= 0 || startLine > endLine {
		return false
	}
	return isValidLine(startLine, totalLines) && isValidLine(endLine, totalLines)
}

func renderSingleLineHighlight(
	w io.Writer,
	line []rune,
	lineNum, startCol, endCol int,
	advanceFn columnAdvanceFn,
) {
	var sourceBuilder strings.Builder
	var markerBuilder strings.Builder

	currentCol := 1
	for _, r := range line {
		nextCol := advanceFn(r, currentCol)
		width := nextCol - currentCol

		sourceBuilder.WriteString(sanitizeRune(r, width))
		markerBuilder.WriteString(buildMarkerFragment(currentCol, startCol, endCol, width))

		currentCol = nextCol
	}

	if currentCol <= startCol {
		markerBuilder.WriteString(strings.Repeat(" ", startCol-currentCol))
		markerBuilder.WriteString("^")
	} else if currentCol < endCol {
		markerBuilder.WriteString(strings.Repeat("~", endCol-currentCol))
	}

	fmt.Fprintf(w, " %4d | %s\n", lineNum, sourceBuilder.String())
	fmt.Fprint(w, "      | ")
	fmt.Fprintln(w, markerBuilder.String())
}

func renderMultiLineHighlight(w io.Writer, lines [][]rune, startL, startC, endL, endC int, advanceFn columnAdvanceFn) {
	firstLine := lines[startL-1]
	renderSingleLineHighlight(w, firstLine, startL, startC, len(firstLine)+1, advanceFn)

	if endL > startL+1 {
		fmt.Fprintln(w, "      | ...")
	}

	lastLine := lines[endL-1]
	renderSingleLineHighlight(w, lastLine, endL, 1, endC, advanceFn)
}

func renderInvalidSpanBlock(w io.Writer, lines [][]rune, startL, startC, endL, endC int, advanceFn columnAdvanceFn) {
	fmt.Fprintln(w, " ── INVALID OR OUT-OF-BOUNDS SPAN DETECTED ────────────────")
	fmt.Fprintf(w, "  Requested: %d:%d to %d:%d | Total lines: %d\n", startL, startC, endL, endC, len(lines))

	lineIdx := determineFallbackLineIndex(startL, len(lines))
	if lineIdx < 0 {
		fmt.Fprintln(w, "(no source available)")
		return
	}

	renderSingleLineHighlight(w, lines[lineIdx], lineIdx+1, startC, endC, advanceFn)
}

func determineFallbackLineIndex(startLine, totalLines int) int {
	if startLine > 0 && startLine <= totalLines {
		return startLine - 1
	}
	if totalLines > 0 {
		return totalLines - 1
	}
	return -1
}

func enforceMinimumColumn(col int) int {
	if col < 1 {
		return 1
	}
	return col
}

func isValidLine(lineNum, totalLines int) bool {
	return lineNum > 0 && lineNum <= totalLines
}

func splitLinesRunes(runes []rune) [][]rune {
	if len(runes) == 0 {
		return nil
	}

	var lines [][]rune
	start := 0

	for i, r := range runes {
		if r == '\n' {
			lines = append(lines, runes[start:i])
			start = i + 1
		}
	}

	if start <= len(runes) {
		lines = append(lines, runes[start:])
	}

	return lines
}

func sanitizeRune(r rune, width int) string {
	if r == '\t' {
		if width < 1 {
			width = 1
		}
		return strings.Repeat(" ", width)
	}
	return string(r)
}

func buildMarkerFragment(currentCol, startCol, endCol, width int) string {
	if currentCol < startCol {
		return strings.Repeat(" ", width)
	}

	if currentCol == startCol {
		if width > 1 && endCol > startCol {
			return "^" + strings.Repeat("~", width-1)
		}
		return "^"
	}

	if currentCol < endCol {
		return strings.Repeat("~", width)
	}

	return ""
}
