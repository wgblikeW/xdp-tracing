package local

import "strings"

type policy struct {
	ds *datastore
}

func newPolicy(ds *datastore) *policy {
	return &policy{ds}
}

func (p *policy) List() (string, error) {
	policies, err := p.ds.db.List()
	return strings.Join(policies, " "), err
}

func (p *policy) Create(policy string) error {
	err := p.ds.db.Append(policy)
	return err
}

func (p *policy) Delete(policy string) error {
	err := p.ds.db.Delete(policy)
	return err
}
