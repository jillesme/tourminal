package render

// Options controls terminal rendering without relying on process-wide state.
// Its zero value renders a light theme with color enabled.
type Options struct {
	Dark    bool
	NoColor bool
}
