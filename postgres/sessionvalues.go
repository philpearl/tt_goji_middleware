package postgres

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"strings"
)

type sessionValues map[string]interface{}

func init() {
	gob.Register(sessionValues{})
	// Register this as well as is used by oAuth code
	gob.Register(map[string]interface{}{})
}

// Value() converts a product into a string suitable for the SQL database
//
// It encodes the data using Gob, then converts to a hex string suitable for the
// bytea type
func (p sessionValues) Value() (driver.Value, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(p)
	if err != nil {
		return nil, err
	}

	return `E'\\x` + hex.EncodeToString(buf.Bytes()) + `'`, nil
}

// Scan() pulls the product out of the SQL database
//
// It expects the data to be in a bytea field, and encoded using GOB
func (p sessionValues) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	b, ok := src.([]uint8)
	if !ok {
		return fmt.Errorf("expected session values to be a byte array")
	}

	bs := string(b)
	if strings.HasPrefix(bs, `E'\\x`) && strings.HasSuffix(bs, `'`) {
		bs = bs[5 : len(bs)-1]

		g, err := hex.DecodeString(bs)
		if err != nil {
			return err
		}
		buf := bytes.NewBuffer(g)
		decoder := gob.NewDecoder(buf)
		return decoder.Decode(&p)
	}
	return fmt.Errorf("session value does not have correct form.  received %s", bs)
}
