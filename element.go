package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Element is a representation of an SVG element.
type Element struct {
	Name       string
	Attributes map[string]string
	Children   []*Element
	Content    string
}

// New creates an Element instance from an SVG input.
func New(source io.Reader) (*Element, error) {
	return decodeFromSource(xml.NewDecoder(source))
}

// Equal checks if two elements are equivalent.
func (e *Element) Equal(o *Element) bool {
	if e.Name != o.Name || e.Content != o.Content ||
		len(e.Attributes) != len(o.Attributes) ||
		len(e.Children) != len(o.Children) {
		return false
	}

	for k, v := range e.Attributes {
		if v != o.Attributes[k] {
			return false
		}
	}

	for i, child := range e.Children {
		if !child.Equal(o.Children[i]) {
			return false
		}
	}
	return true
}

// deserialize creates element from decoder token.
func deserialize(token xml.StartElement) *Element {
	element := &Element{
		Name:       token.Name.Local,
		Attributes: map[string]string{},
	}

	for _, attr := range token.Attr {
		element.Attributes[attr.Name.Local] = attr.Value
	}

	return element
}

// decodeFromSource creates the first element from the decoder.
func decodeFromSource(decoder *xml.Decoder) (*Element, error) {
	var root *Element

	for {
		token, err := decoder.Token()
		if token == nil && err == io.EOF {
			return root, nil

		} else if err != nil {
			return nil, fmt.Errorf("Error decoding element: %s", err)
		}

		if element, found := token.(xml.StartElement); found {
			root = deserialize(element)
			break
		}
	}

	if err := decode(root, decoder); err != nil && err != io.EOF {
		return nil, fmt.Errorf("Error decoding element: %s", err)
	}

	return root, nil
}

// decode decodes the child elements of element.
func decode(e *Element, decoder *xml.Decoder) error {
	for {
		token, err := decoder.Token()
		if token == nil && err == io.EOF {
			break

		} else if err != nil {
			return err
		}

		switch element := token.(type) {
		case xml.StartElement:
			nextElement := deserialize(element)
			if err := decode(nextElement, decoder); err != nil {
				return err
			}

			e.Children = append(e.Children, nextElement)

		case xml.CharData:
			data := strings.TrimSpace(string(element))
			if data != "" {
				e.Content = string(element)
			}

		case xml.EndElement:
			if element.Name.Local == e.Name {
				return nil
			}
		}
	}

	return nil
}
