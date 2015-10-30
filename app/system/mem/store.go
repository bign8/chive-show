package mem

type memStore struct {
	data map[string]string
}

func (store *memStore) Create() error {
	return nil
}

func (store *memStore) Read() error {
	return nil
}

func (store *memStore) Update() error {
	return nil
}

func (store *memStore) Delete() error {
	return nil
}
