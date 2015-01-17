package omap

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// An OrderedMap keeps track of the order keys are placed in it
//
// Rather than be recursive, the values are of type json.RawMessage
// and thus can be decoded separately
type OrderedMap struct {
	m   map[string]json.RawMessage
	k   []string
	has map[string]bool
}

type keyVal struct {
	Key string
	Val json.RawMessage
}

// NewOrderedMap returns a new OrderedMap
func NewOrderedMap() *OrderedMap {
	o := new(OrderedMap)
	o.m = make(map[string]json.RawMessage)
	o.k = make([]string, 0)
	o.has = make(map[string]bool)
	return o
}

// Set sets the value of a key
//
// If the key already exists in the map it will retain its position
// otherwise it will be placed at the end
func (o *OrderedMap) Set(k string, v json.RawMessage) {
	if !o.has[k] {
		o.has[k] = true
		o.k = append(o.k, k)
	}
	o.m[k] = v
}

// Get returns the json value associated with a key
func (o *OrderedMap) Get(k string) json.RawMessage {
	return o.m[k]
}

func nextPair(r io.Reader) (pair *keyVal, rest io.Reader, err error) {
	//next token should be '{', ',', or '}'
	var b = make([]byte, 1)
	for {
		_, err = r.Read(b)
		if err != nil {
			return
		}
		switch b[0] {
		case '\t', '\n', '\r', ' ': //skip whitespace
			continue
		case '{', ',': //new object, or next key
			goto readKey
		case '}': //all done
			return nil, nil, nil
		default:
			//TODO: add beter info/error type
			return nil, nil, errors.New("Invalid token: expected '{' or ',' but found '" + string(b[0]) + "'")
		}
	}
readKey:
	var key string
	d := json.NewDecoder(r)
	err = d.Decode(&key)
	if err != nil {
		return
	}
	r = d.Buffered()
	for {
		_, err = r.Read(b)
		if err != nil {
			return
		}
		switch b[0] {
		case '\t', '\n', '\r', ' ': //skip whitespace
			continue
		case ':': //key/val delimiter
			goto readVal
		default:
			//TODO: add beter info/error type
			return nil, nil, errors.New("Invalid token: expected ':' but found '" + string(b) + "'")
		}
	}
readVal:
	d = json.NewDecoder(r)
	var val json.RawMessage
	err = d.Decode(&val)
	if err != nil {
		return
	}
	return &keyVal{key, val}, d.Buffered(), nil
}

func (o *OrderedMap) UnmarshalJSON(b []byte) error {
	var r io.Reader
	var pair *keyVal
	var err error
	r = bytes.NewReader(b)
	for {
		pair, r, err = nextPair(r)
		if err != nil {
			return err
		}
		if pair != nil {
			o.Set(pair.Key, pair.Val)
		} else {
			break
		}
	}
	return nil
}
func (o *OrderedMap) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteByte('{')
	for i, v := range o.k {
		if i > 0 {
			b.WriteByte(',')
		}
		keyData, err := json.Marshal(&v)
		if err != nil {
			return nil, err
		}
		b.Write(keyData)
		b.WriteByte(':')
		b.Write([]byte(o.m[v]))
	}
	b.WriteByte('}')

	return b.Bytes(), nil
}
