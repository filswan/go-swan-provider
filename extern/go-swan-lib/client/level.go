package client

import (
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"

	"github.com/syndtr/goleveldb/leveldb"
)

func LevelDbPut(dbFilepath, key string, value interface{}) error {
	db, err := leveldb.OpenFile(dbFilepath, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer db.Close()

	var valStr string
	switch valType := value.(type) {
	case string:
		valStr = valType
		logs.GetLogger().Info("this is already string")
	default:
		valStr = utils.ToJson(value)
	}

	err = db.Put([]byte(key), []byte(valStr), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func LevelDbGet(dbFilepath, key string) ([]byte, error) {
	db, err := leveldb.OpenFile(dbFilepath, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	defer db.Close()

	data, err := db.Get([]byte(key), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return data, nil
}

func LevelDbDelete(dbFilepath, key, value string) error {
	db, err := leveldb.OpenFile(dbFilepath, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer db.Close()

	err = db.Delete([]byte("key"), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}
