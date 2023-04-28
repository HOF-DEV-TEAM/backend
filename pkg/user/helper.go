package user

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func EncodeString(s string) string {
	data := base64.StdEncoding.EncodeToString([]byte(s))
	return string(data)
}

func DecodeString(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type OtpGenerator interface {
	Generate() string
}

type otpGenerator struct{}

func NewOTPGenerator() OtpGenerator {
	rand.Seed(time.Now().UnixNano())
	return &otpGenerator{}
}

func (*otpGenerator) Generate() string {
	return fmt.Sprintf("%06d", rand.Intn(999999))
}

func remove(slice []FavBody, s int) []FavBody {
	return append(slice[:s], slice[s+1:]...)
}

func SaveFavouritesJSONBValue(f []FavBody) (driver.Value, error) {
	j, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return driver.Value([]byte(j)), nil
}

type ScanFavourite []FavBody

func (f *ScanFavourite) Scan(value interface{}) error {
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

func SaveDevicesJSONBValue(d []Device) (driver.Value, error) {
	j, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return driver.Value([]byte(j)), nil
}

type ScanDevices []Device

func (d *ScanDevices) Scan(value interface{}) error {
	var source []byte
	switch value.(type) {
	case []uint8:
		source = []byte(value.([]uint8))
	case nil:
		return nil
	default:
		return errors.New("type assertion to []byte failed")
	}
	err := json.Unmarshal(source, &d)
	if err != nil {
		return err
	}
	return nil
}

type ScanDevice Device

func (d *ScanDevice) Scan(value interface{}) error {
	var source []byte
	switch value.(type) {
	case []uint8:
		source = []byte(value.([]uint8))
	case nil:
		return nil
	default:
		return errors.New("type assertion to []byte failed")
	}
	err := json.Unmarshal(source, &d)
	if err != nil {
		return err
	}
	return nil
}
