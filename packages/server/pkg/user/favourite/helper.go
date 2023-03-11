package favourite

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

func remove(slice []FavBody, s int) []FavBody {
	return append(slice[:s], slice[s+1:]...)
}

func Value(f []FavBody) (driver.Value, error) {
	j, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return driver.Value([]byte(j)), nil
}

type Favourite []FavBody

func (f *Favourite) Scan(value interface{}) error {
	var source []byte
	switch value.(type) {
	case []uint8:
		source = []byte(value.([]uint8))
	case nil:
		return nil
	default:
		return errors.New("type assertion to []byte failed")
	}
	err := json.Unmarshal(source, &f)
	if err != nil {
		return err
	}
	return nil
}
