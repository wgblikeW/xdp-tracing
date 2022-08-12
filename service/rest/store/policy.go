package store

// PolicyStore defines the policy storage interface.
type PolicyStore interface {
	List() (string, error)
	Create(string) error
	Delete(string) error
}
