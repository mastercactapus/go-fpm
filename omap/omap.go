package omap

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
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

type decodeReader struct {
	r         *bytes.Reader
	readBytes int
}

func (d *decodeReader) Read(b []byte) (int, error) {
	l, err := d.r.Read(b)
	d.readBytes += l
	return l, err
}
func (d *decodeReader) ReadByte() (byte, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		return 0, err
	}
	d.readBytes++
	return b, nil
}
func (d *decodeReader) UnreadReader(r io.Reader) error {
	dat, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	_, err = d.r.Seek(int64(-len(dat)), 1)
	if err != nil {
		return err
	}
	return nil
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

func nextPair(d *decodeReader) (pair *keyVal, err error) {

	//next token should be '{', ',', or '}'
	for {
		b, err := d.ReadByte()
		if err != nil {
			return nil, err
		}
		switch b {
		case '\t', '\n', '\r', ' ': //skip whitespace
			continue
		case '{', ',': //new object, or next key
			goto readKey
		case '}': //all done
			return nil, nil
		default:
			//TODO: add beter info/error type
			return nil, errors.New("Invalid token: expected '{' or ',' but found '" + string(b) + "'")
		}
	}
readKey:
	var key string
	dec := json.NewDecoder(d)
	err = dec.Decode(&key)
	if err != nil {
		return
	}
	d.UnreadReader(dec.Buffered())

	for {
		b, err := d.ReadByte()
		if err != nil {
			return nil, err
		}
		switch b {
		case '\t', '\n', '\r', ' ': //skip whitespace
			continue
		case ':': //key/val delimiter
			goto readVal
		default:
			//TODO: add beter info/error type
			return nil, errors.New("Invalid token: expected ':' but found '" + string(b) + "'")
		}
	}
readVal:
	dec = json.NewDecoder(d)
	var val json.RawMessage
	err = dec.Decode(&val)
	if err != nil {
		return
	}
	d.UnreadReader(dec.Buffered())
	return &keyVal{key, val}, nil
}

func (o *OrderedMap) UnmarshalJSON(b []byte) error {
	var pair *keyVal
	var err error
	d := &decodeReader{r: bytes.NewReader(b)}

	for {
		pair, err = nextPair(d)
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
