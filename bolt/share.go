package bolt

import (
	"github.com/asdine/storm"
	fm "github.com/hacdias/filemanager"
)

type ShareStore struct {
	DB *storm.DB
}

func (s ShareStore) Get(hash string) (*fm.ShareLink, error) {
	var v *fm.ShareLink
	err := s.DB.One("Hash", hash, &v)
	return v, err
}

func (s ShareStore) GetByPath(hash string) ([]*fm.ShareLink, error) {
	var v []*fm.ShareLink
	err := s.DB.Find("Path", hash, &v)
	return v, err
}

func (s ShareStore) Gets(hash string) ([]*fm.ShareLink, error) {
	var v []*fm.ShareLink
	err := s.DB.All(&v)
	return v, err
}

func (s ShareStore) Save(l *fm.ShareLink) error {
	return s.DB.Save(l)
}

func (s ShareStore) Delete(hash string) error {
	return s.DB.DeleteStruct(&fm.ShareLink{Hash: hash})
}
