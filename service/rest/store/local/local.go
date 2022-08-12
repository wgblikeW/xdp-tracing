package local

import (
	"github.com/p1nant0m/xdp-tracing/pkg/db"
	"github.com/p1nant0m/xdp-tracing/service/rest/store"
)

type datastore struct {
	db *db.LocalStorage
}

func (ds *datastore) Policy() store.PolicyStore {
	return newPolicy(ds)
}

var localStorageFactory store.Factory

func GetLocalStorageFactoryOr() (store.Factory, error) {
	if localStorageFactory != nil {
		return localStorageFactory, nil
	}

	dbIns, _ := db.NewLocalStorage()
	localStorageFactory = &datastore{dbIns}

	return localStorageFactory, nil
}
