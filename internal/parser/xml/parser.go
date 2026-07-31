package xml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/marcuwynu23/dataflute/internal/parser/scalar"
)

// Parser decodes XML documents into generic, JSON-friendly values.
type Parser struct{}

// New returns a new XML parser.
func New() *Parser { return &Parser{} }

// Extensions returns the file extensions handled by this parser.
func (*Parser) Extensions() []string { return []string{".xml"} }

type xmlAttr struct {
	name  string
	value string
}

type xmlElement struct {
	name     string
	attrs    []xmlAttr
	children []*xmlElement
	text     strings.Builder
}

// Parse decodes the XML document at path.
func (*Parser) Parse(path string) (any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	root, err := parseRoot(xml.NewDecoder(f))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return root.rootValue(), nil
}

func parseRoot(dec *xml.Decoder) (*xmlElement, error) {
	var root *xmlElement
	stack := make([]*xmlElement, 0, 8)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			e := &xmlElement{name: t.Name.Local}
			for _, a := range t.Attr {
				e.attrs = append(e.attrs, xmlAttr{name: a.Name.Local, value: a.Value})
			}
			if len(stack) > 0 {
				stack[len(stack)-1].children = append(stack[len(stack)-1].children, e)
			} else if root == nil {
				root = e
			} else {
				return nil, errors.New("multiple root elements")
			}
			stack = append(stack, e)
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, errors.New("unexpected closing tag")
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].text.WriteString(string(t))
			}
		}
	}
	if len(stack) > 0 {
		return nil, errors.New("unclosed element")
	}
	if root == nil {
		return nil, errors.New("empty document")
	}
	return root, nil
}

// rootValue unwraps a root element that contains a single repeated child
// tag into an array of records, mirroring how CSV datasets behave.
func (e *xmlElement) rootValue() any {
	names := map[string]struct{}{}
	for _, c := range e.children {
		names[c.name] = struct{}{}
	}
	if len(e.attrs) == 0 && len(names) == 1 && len(e.children) > 0 {
		out := make([]any, 0, len(e.children))
		for _, c := range e.children {
			out = append(out, c.value())
		}
		return out
	}
	return e.value()
}

// value converts an element into a generic value. Attributes are exposed
// with an "@" prefix; mixed text content is exposed as "#text".
func (e *xmlElement) value() any {
	children := map[string][]any{}
	for _, c := range e.children {
		children[c.name] = append(children[c.name], c.value())
	}
	text := strings.TrimSpace(e.text.String())

	if len(children) == 0 && len(e.attrs) == 0 {
		if text == "" {
			return nil
		}
		return scalar.Value(text)
	}

	out := make(map[string]any, len(children)+len(e.attrs)+1)
	for _, a := range e.attrs {
		out["@"+a.name] = a.value
	}
	for name, vals := range children {
		if len(vals) == 1 {
			out[name] = vals[0]
		} else {
			out[name] = vals
		}
	}
	if text != "" {
		out["#text"] = text
	}
	return out
}
