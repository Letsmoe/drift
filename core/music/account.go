package music

// Account is the YouTube Music account a library import reads from.
type Account struct {
	Name string `json:"name"`
	// Handle is the channel handle ("@name"); may be empty.
	Handle string `json:"handle"`
	// Email is set when YouTube shows it in the account menu; may be empty.
	Email    string `json:"email"`
	PhotoURL string `json:"photoUrl"`
}
