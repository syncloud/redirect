package mailin

// Route is where a recipient's mail has to go: the device's name, which the
// multiplexer uses to pick the right tunnel, and the multiplexer to ask.
type Route struct {
	Domain string
	Muxer  string
}
