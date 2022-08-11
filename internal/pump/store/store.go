package store

var client Factory

type Factory interface {
	Session() SessionStore
	Sessions() SessionsStore
}

func Client() Factory {
	return client
}

func SetClient(factory Factory) {
	client = factory
}
