package svg_test

import (
	"bytes"
	"strings"
	"testing"

	. "github.com/catiepg/svg"
)

func TestElementNew(t *testing.T) {
	tests := []struct {
		description string
		raw         string
		expected    *Element
	}{
		{
			description: "simple element",
			raw: `
			<svg width="100" height="100">
				<circle cx="50" cy="50" />
			</svg>
			`,
			expected: &Element{
				Name: "svg",
				Attributes: map[string]string{
					"width":  "100",
					"height": "100",
				},
				Children: []*Element{
					{
						Name:       "circle",
						Attributes: map[string]string{"cx": "50", "cy": "50"},
						Children:   []*Element{},
					},
				},
			},
		},
		{
			description: "nested element",
			raw: `
			<svg height="400" width="450">
				<g stroke="black" stroke-width="3">
					<path d="M 10 20 L 15 -25" />
					<path d="M 25 50 L 15 30" />
				</g>
			</svg>
			`,
			expected: &Element{
				Name: "svg",
				Attributes: map[string]string{
					"width":  "450",
					"height": "400",
				},
				Children: []*Element{
					{
						Name: "g",
						Attributes: map[string]string{
							"stroke":       "black",
							"stroke-width": "3",
						},
						Children: []*Element{
							{
								Name: "path",
								Attributes: map[string]string{
									"d": "M 10 20 L 15 -25",
								},
								Children: []*Element{},
							},
							{
								Name: "path",
								Attributes: map[string]string{
									"d": "M 25 50 L 15 30",
								},
								Children: []*Element{},
							},
						},
					},
				},
			},
		},
		{
			description: "element with text",
			raw: `
			<svg width="100" height="100">
				<text>Hello</text>
			</svg>
			`,
			expected: &Element{
				Name: "svg",
				Attributes: map[string]string{
					"width":  "100",
					"height": "100",
				},
				Children: []*Element{
					{
						Name:    "text",
						Content: "Hello",
					},
				},
			},
		},
		{
			description: "element with empty text child",
			raw:         "<svg><text>\t\n</text></svg>",
			expected: &Element{
				Name: "svg",
				Children: []*Element{
					{
						Name:    "text",
						Content: "",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual, err := New(strings.NewReader(test.raw))
			if err != nil {
				t.Fatalf("New: unexpected error: %s", err)
			}

			if !test.expected.Equal(actual) {
				t.Fatalf("New: expected %v, actual %v", test.expected, actual)
			}
		})
	}
}

func TestElementNewErrors(t *testing.T) {
	tests := []struct {
		description    string
		raw            string
		expectedPrefix string
	}{
		{
			description:    "simple element",
			raw:            `<svg`,
			expectedPrefix: "Error decoding element",
		},
		{
			description:    "nested element",
			raw:            `<svg> <a </svg>`,
			expectedPrefix: "Error decoding element",
		},
		{
			description:    "sibling element",
			raw:            `<svg> <a> <b </a> </svg>`,
			expectedPrefix: "Error decoding element",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual, err := New(strings.NewReader(test.raw))
			if actual != nil {
				t.Fatalf("New: expected element to be nil, actual: %s", actual)
			}

			if !strings.HasPrefix(err.Error(), test.expectedPrefix) {
				t.Fatalf("New: expected error to have prefix '%s', actual '%s'",
					test.expectedPrefix, err)
			}
		})
	}
}

func TestElementNewEmpty(t *testing.T) {
	actual, err := New(strings.NewReader(""))
	if err != nil {
		t.Fatalf("New: unexpected error: %s", err)
	}

	if actual != nil {
		t.Errorf("New: expected nil, actual %v", actual)
	}
}

func TestElementRender(t *testing.T) {
	tests := []struct {
		description string
		element     *Element
		expected    string
	}{
		{
			description: "simple element",
			element: &Element{
				Name:       "svg",
				Attributes: map[string]string{"fill": "blue"},
			},
			expected: `<svg fill="blue"></svg>`,
		},
		{
			description: "nested element",
			element: &Element{
				Name:       "g",
				Attributes: map[string]string{"stroke": "black"},
				Children: []*Element{
					{Name: "path", Attributes: map[string]string{"d": "m 1 2"}},
					{Name: "path", Attributes: map[string]string{"d": "m 3 4"}},
				},
			},
			expected: `<g stroke="black"><path d="m 1 2"></path><path d="m 3 4"></path></g>`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			buf := &bytes.Buffer{}
			if err := test.element.Render(buf); err != nil {
				t.Fatalf("Render: unexpected error: %s", err)
			}

			if actual := buf.String(); test.expected != actual {
				t.Fatalf("Render: expected %s, actual %s",
					test.expected, buf.String())
			}
		})
	}
}

func TestElementRenderErrors(t *testing.T) {
	tests := []struct {
		description    string
		element        *Element
		expectedPrefix string
	}{
		{
			description:    "empty element",
			element:        &Element{},
			expectedPrefix: "Could not render element",
		},
		{
			description: "child with no name",
			element: &Element{
				Name: "g",
				Children: []*Element{
					{Attributes: map[string]string{"fill": "black"}},
				},
			},
			expectedPrefix: "Could not render element",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.element.Render(&bytes.Buffer{})
			if err == nil {
				t.Fatalf("Render: expected error, actual nil")
			}

			if !strings.HasPrefix(err.Error(), test.expectedPrefix) {
				t.Fatalf("Render: expected prefix %s, actual %v",
					test.expectedPrefix, err)
			}
		})
	}
}

func TestElementEqual(t *testing.T) {
	tests := []struct {
		description string
		element     *Element
		other       *Element
		expected    bool
	}{
		{
			description: "empty",
			element:     &Element{},
			other:       &Element{},
			expected:    true,
		},
		{
			description: "deep equal",
			element: &Element{
				Name:       "svg",
				Attributes: map[string]string{"width": "100", "height": "100"},
				Children: []*Element{
					{
						Name:       "circle",
						Attributes: map[string]string{"cx": "50", "cy": "50"},
						Children:   []*Element{},
					},
				},
			},
			other: &Element{
				Name:       "svg",
				Attributes: map[string]string{"width": "100", "height": "100"},
				Children: []*Element{
					{
						Name:       "circle",
						Attributes: map[string]string{"cx": "50", "cy": "50"},
						Children:   []*Element{},
					},
				},
			},
			expected: true,
		},
		{
			description: "different names",
			element: &Element{
				Name: "svg",
			},
			other: &Element{
				Name: "rect",
			},
			expected: false,
		},
		{
			description: "different attributes",
			element: &Element{
				Attributes: map[string]string{"width": "100", "height": "100"},
			},
			other: &Element{
				Attributes: map[string]string{"width": "1000", "fill": "white"},
			},
			expected: false,
		},
		{
			description: "different children",
			element: &Element{
				Children: []*Element{
					{Name: "circle"},
				},
			},
			other: &Element{
				Children: []*Element{
					{Name: "rect"},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := test.element.Equal(test.other)
			if actual != test.expected {
				t.Fatalf("Element: expected %v, actual %v", test.expected, actual)
			}
		})
	}
}
