package lazynode

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
)

const (
	eRaw = iota
	eDoc
	eAry
)

var (
	ErrUnknownType  = errors.New("unknown object type")
	ErrInvalid      = errors.New("invalid state detected")
	ErrInvalidIndex = errors.New("invalid index referenced")
	ErrMissing      = errors.New("missing value")
)

type LazyNode struct {
	raw   *json.RawMessage
	doc   PartialDoc
	ary   PartialArray
	which int
}

type (
	PartialDoc   map[string]*LazyNode
	PartialArray []*LazyNode
)

func NewLazyNode(raw *json.RawMessage) *LazyNode {
	return &LazyNode{raw: raw, doc: nil, ary: nil, which: eRaw}
}

func (n *LazyNode) MarshalJSON() ([]byte, error) {
	switch n.which {
	case eRaw:
		return json.Marshal(n.raw)
	case eDoc:
		return json.Marshal(n.doc)
	case eAry:
		return json.Marshal(n.ary)
	default:
		return nil, ErrUnknownType
	}
}

func (n *LazyNode) UnmarshalJSON(data []byte) error {
	dest := make(json.RawMessage, len(data))
	copy(dest, data)
	n.raw = &dest
	n.which = eRaw
	return nil
}

func (n *LazyNode) Which() int {
	return n.which
}

func (n *LazyNode) SetWhich(w int) {
	n.which = w
}

func (n *LazyNode) GetDoc() PartialDoc {
	return n.doc
}

func (n *LazyNode) SetDoc(pDoc PartialDoc) {
	n.doc = pDoc
}

func (n *LazyNode) GetArray() PartialArray {
	return n.ary
}

func (n *LazyNode) SetArray(pArray PartialArray) {
	n.ary = pArray
}

func (n *LazyNode) IsRawEmpty() bool {
	return n.raw == nil
}

func (n *LazyNode) IsArray() bool {
Loop:
	for _, c := range *n.raw {
		switch c {
		case ' ':
		case '\n':
		case '\t':
			continue
		case '[':
			return true
		default:
			break Loop
		}
	}

	return false
}

func (n *LazyNode) IntoDoc() (*PartialDoc, error) {
	if n.which == eDoc {
		return &n.doc, nil
	}

	if n.raw == nil {
		return nil, ErrInvalid
	}

	err := json.Unmarshal(*n.raw, &n.doc)
	if err != nil {
		return nil, err
	}

	n.which = eDoc
	return &n.doc, nil
}

func (n *LazyNode) IntoAry() (*PartialArray, error) {
	if n.which == eAry {
		return &n.ary, nil
	}

	if n.raw == nil {
		return nil, ErrInvalid
	}

	err := json.Unmarshal(*n.raw, &n.ary)
	if err != nil {
		return nil, err
	}

	n.which = eAry
	return &n.ary, nil
}

func (n *LazyNode) compact() []byte {
	buf := &bytes.Buffer{}

	if n.raw == nil {
		return nil
	}

	err := json.Compact(buf, *n.raw)
	if err != nil {
		return *n.raw
	}

	return buf.Bytes()
}

func (n *LazyNode) TryDoc() bool {
	if n.raw == nil {
		return false
	}

	err := json.Unmarshal(*n.raw, &n.doc)
	if err != nil {
		return false
	}

	n.which = eDoc
	return true
}

func (n *LazyNode) TryAry() bool {
	if n.raw == nil {
		return false
	}

	err := json.Unmarshal(*n.raw, &n.ary)
	if err != nil {
		return false
	}

	n.which = eAry
	return true
}

func (n *LazyNode) Equal(o *LazyNode) bool {
	if n.which == eRaw {
		if !n.TryDoc() && !n.TryAry() {
			if o.which != eRaw {
				return false
			}

			return bytes.Equal(n.compact(), o.compact())
		}
	}

	if n.which == eDoc {
		if o.which == eRaw {
			if !o.TryDoc() {
				return false
			}
		}

		if o.which != eDoc {
			return false
		}

		if len(n.doc) != len(o.doc) {
			return false
		}

		for k, v := range n.doc {
			ov, ok := o.doc[k]

			if !ok {
				return false
			}

			if (v == nil) != (ov == nil) {
				return false
			}

			if v == nil && ov == nil {
				continue
			}

			if !v.Equal(ov) {
				return false
			}
		}

		return true
	}

	if o.which != eAry && !o.TryAry() {
		return false
	}

	if len(n.ary) != len(o.ary) {
		return false
	}

	for idx, val := range n.ary {
		if !val.Equal(o.ary[idx]) {
			return false
		}
	}

	return true
}

func (d *PartialDoc) Set(key string, val *LazyNode) error {
	(*d)[key] = val
	return nil
}

func (d *PartialDoc) Add(key string, val *LazyNode) error {
	(*d)[key] = val
	return nil
}

func (d *PartialDoc) Get(key string) (*LazyNode, error) {
	return (*d)[key], nil
}

func (d *PartialDoc) Remove(key string) error {
	_, ok := (*d)[key]
	if !ok {
		return errors.Wrapf(ErrMissing, "Unable to remove nonexistent key: %s", key)
	}

	delete(*d, key)
	return nil
}

// set should only be used to implement the "replace" operation, so "key" must
// be an already existing index in "d".
func (d *PartialArray) Set(key string, val *LazyNode) error {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return err
	}

	if idx < 0 {
		if idx < -len(*d) {
			return errors.Wrapf(ErrInvalidIndex, "Unable to access invalid index: %d", idx)
		}
		idx += len(*d)
	}

	(*d)[idx] = val
	return nil
}

func (d *PartialArray) Add(key string, val *LazyNode) error {
	if key == "-" {
		*d = append(*d, val)
		return nil
	}

	idx, err := strconv.Atoi(key)
	if err != nil {
		return errors.Wrapf(err, "value was not a proper array index: '%s'", key)
	}

	sz := len(*d) + 1

	ary := make([]*LazyNode, sz)

	cur := *d

	if idx >= len(ary) {
		return errors.Wrapf(ErrInvalidIndex, "Unable to access invalid index: %d", idx)
	}

	if idx < 0 {
		if idx < -len(ary) {
			return errors.Wrapf(ErrInvalidIndex, "Unable to access invalid index: %d", idx)
		}
		idx += len(ary)
	}

	copy(ary[0:idx], cur[0:idx])
	ary[idx] = val
	copy(ary[idx+1:], cur[idx:])

	*d = ary
	return nil
}

func (d *PartialArray) Get(key string) (*LazyNode, error) {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return nil, err
	}

	if idx < 0 {
		if idx < -len(*d) {
			return nil, errors.Wrapf(ErrInvalidIndex, "Unable to access invalid index: %d", idx)
		}
		idx += len(*d)
	}

	if idx >= len(*d) {
		return nil, errors.Wrapf(ErrInvalidIndex, "Unable to access invalid index: %d", idx)
	}

	return (*d)[idx], nil
}

func (d *PartialArray) Remove(key string) error {
	idx, err := strconv.Atoi(key)
	if err != nil {
		return err
	}

	cur := *d

	if idx >= len(cur) {
		return errors.Wrapf(ErrInvalidIndex, "Unable to access invalid index: %d", idx)
	}

	if idx < 0 {
		if idx < -len(cur) {
			return errors.Wrapf(ErrInvalidIndex, "Unable to access invalid index: %d", idx)
		}
		idx += len(cur)
	}

	ary := make([]*LazyNode, len(cur)-1)

	copy(ary[0:idx], cur[0:idx])
	copy(ary[idx:], cur[idx+1:])

	*d = ary
	return nil
}

func DeepCopy(src *LazyNode) (*LazyNode, int, error) {
	if src == nil {
		return nil, 0, nil
	}
	a, err := src.MarshalJSON()
	if err != nil {
		return nil, 0, err
	}
	sz := len(a)
	ra := make(json.RawMessage, sz)
	copy(ra, a)
	return NewLazyNode(&ra), sz, nil
}
