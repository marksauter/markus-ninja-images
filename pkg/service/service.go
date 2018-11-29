package service

import (
	"github.com/marksauter/markus-ninja-images/pkg/myconf"
	"github.com/marksauter/markus-ninja-images/pkg/mylog"
	"github.com/marksauter/markus-ninja-images/pkg/util"
)

type Services struct {
	Storage *StorageService
}

func NewServices(conf *myconf.Config) (*Services, error) {
	storageSvc, err := NewStorageService(conf)
	if err != nil {
		mylog.Log.WithError(err).Error(util.Trace(""))
		return nil, err
	}
	return &Services{
		Storage: storageSvc,
	}, nil
}
