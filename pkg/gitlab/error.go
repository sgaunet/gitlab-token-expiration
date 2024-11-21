package gitlab

// ErrorMessage is a struct that contains an error message
// It is returned by the Gitlab API when an error occurs
type ErrorMessage struct {
	Message string `json:"message"`
}
