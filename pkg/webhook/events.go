package webhook

// Event name constants. These are the only valid event names — not user-configurable.
const (
	EventPdfJobQueued    = "PdfJobQueued"
	EventPdfJobCompleted = "PdfJobCompleted"
	EventPdfJobFailed    = "PdfJobFailed"
)

// AllEvents returns the list of all supported event names for API responses.
func AllEvents() []string {
	return []string{EventPdfJobQueued, EventPdfJobCompleted, EventPdfJobFailed}
}
