package bot

// Feature interface for modular bot functionality
type Feature interface {
	Name() string
	Execute(session *SessionChrome, args map[string]interface{}) (interface{}, error)
}
