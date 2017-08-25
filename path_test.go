package svg_test

import (
	"fmt"
	"testing"

	. "github.com/catiepg/svg"
)

func TestPathCommandEqual(t *testing.T) {
	tests := []struct {
		description string
		command     *PathCommand
		other       *PathCommand
		expected    bool
	}{
		{
			description: "empty",
			command:     &PathCommand{},
			other:       &PathCommand{},
			expected:    true,
		},
		{
			description: "equal",
			command: &PathCommand{
				Symbol: "m",
				Params: []float64{1, 2, 3},
			},
			other: &PathCommand{
				Symbol: "m",
				Params: []float64{1, 2, 3},
			},
			expected: true,
		},
		{
			description: "different symbols",
			command: &PathCommand{
				Symbol: "m",
			},
			other: &PathCommand{
				Symbol: "l",
			},
			expected: false,
		},
		{
			description: "different params",
			command: &PathCommand{
				Params: []float64{1, 2, 3},
			},
			other: &PathCommand{
				Params: []float64{2, 3, 5},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := test.command.Equal(test.other)
			if actual != test.expected {
				t.Fatalf("Path: expected %v, actual %v", test.expected, actual)
			}
		})
	}
}

func TestPathEqual(t *testing.T) {
	tests := []struct {
		description string
		path        *Path
		other       *Path
		expected    bool
	}{
		{
			description: "equal",
			path: &Path{
				Commands: []*PathCommand{
					{Symbol: "m", Params: []float64{1, 2}},
				},
			},
			other: &Path{
				Commands: []*PathCommand{
					{Symbol: "m", Params: []float64{1, 2}},
				},
			},
			expected: true,
		},
		{
			description: "different number of commands",
			path: &Path{
				Commands: []*PathCommand{
					{Symbol: "m"},
				},
			},
			other: &Path{
				Commands: []*PathCommand{
					{Symbol: "m"}, {Symbol: "l"},
				},
			},
			expected: false,
		},
		{
			description: "different commands",
			path: &Path{
				Commands: []*PathCommand{
					{Symbol: "l", Params: []float64{1, 2}},
				},
			},
			other: &Path{
				Commands: []*PathCommand{
					{Symbol: "m", Params: []float64{1, 2}},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := test.path.Equal(test.other)
			if actual != test.expected {
				t.Fatalf("Path: expected %v, actual %v", test.expected, actual)
			}
		})
	}
}

func TestPathParser(t *testing.T) {
	tests := []struct {
		description string
		rawPath     string
		expected    []*PathCommand
	}{
		{
			description: "simple path",
			rawPath:     "M 10,20 L 30,30 Z",
			expected: []*PathCommand{
				{Symbol: "M", Params: []float64{10, 20}},
				{Symbol: "L", Params: []float64{30, 30}},
				{Symbol: "Z", Params: []float64{}},
			},
		},
		{
			description: "path with decimal values with no numbers on the left",
			rawPath:     "M .2.3 L 30,30 Z",
			expected: []*PathCommand{
				{Symbol: "M", Params: []float64{0.2, 0.3}},
				{Symbol: "L", Params: []float64{30, 30}},
				{Symbol: "Z", Params: []float64{}},
			},
		},
		{
			description: "path with decimal values",
			rawPath:     "M 0.2 1.3 L 30,30 Z",
			expected: []*PathCommand{
				{Symbol: "M", Params: []float64{0.2, 1.3}},
				{Symbol: "L", Params: []float64{30, 30}},
				{Symbol: "Z", Params: []float64{}},
			},
		},
		{
			description: "path with negative values, no space",
			rawPath:     "M10-20 L30,30Z",
			expected: []*PathCommand{
				{Symbol: "M", Params: []float64{10, -20}},
				{Symbol: "L", Params: []float64{30, 30}},
				{Symbol: "Z", Params: []float64{}},
			},
		},
		{
			description: "path with negative values",
			rawPath:     "M 10-20 L 30,30 L 40,40 Z",
			expected: []*PathCommand{
				{Symbol: "M", Params: []float64{10, -20}},
				{Symbol: "L", Params: []float64{30, 30}},
				{Symbol: "L", Params: []float64{40, 40}},
				{Symbol: "Z", Params: []float64{}},
			},
		},
		{
			description: "path without end symbol",
			rawPath:     "M10,20 L20,30 L10,20",
			expected: []*PathCommand{
				{Symbol: "M", Params: []float64{10, 20}},
				{Symbol: "L", Params: []float64{20, 30}},
				{Symbol: "L", Params: []float64{10, 20}},
			},
		},
		{
			description: "path with multiple end symbols",
			rawPath:     " M 10,20 30,40 Z l 10,20 Z",
			expected: []*PathCommand{
				{Symbol: "M", Params: []float64{10, 20}},
				{Symbol: "L", Params: []float64{30, 40}},
				{Symbol: "Z", Params: []float64{}},
				{Symbol: "l", Params: []float64{10, 20}},
				{Symbol: "Z", Params: []float64{}},
			},
		},
		{
			description: "scientific notation",
			rawPath:     " M 10,200e-2",
			expected: []*PathCommand{
				{Symbol: "M", Params: []float64{10, 2}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			path, err := NewPath(test.rawPath)
			if err != nil {
				t.Fatalf("Path: unexpected error: %v", err)
			}

			for i, command := range test.expected {
				if !command.Equal(path.Commands[i]) {
					t.Errorf("Path: expected %v, actual %v",
						command, path.Commands[i])
				}
			}
		})
	}
}

func TestPathImplicitLinetoCommands(t *testing.T) {
	path, err := NewPath("M 10,20 30,40 Z m 10,20 30,40 Z")
	if err != nil {
		t.Fatalf("Path: unexpected error: %v", err)
	}

	expectedCommands := []*PathCommand{
		{Symbol: "M", Params: []float64{10, 20}},
		{Symbol: "L", Params: []float64{30, 40}},
		{Symbol: "Z", Params: []float64{}},
		{Symbol: "m", Params: []float64{10, 20}},
		{Symbol: "l", Params: []float64{30, 40}},
		{Symbol: "Z", Params: []float64{}},
	}

	for i, command := range expectedCommands {
		if !command.Equal(path.Commands[i]) {
			t.Errorf("Path: expected %v, actual %v", command, path.Commands[i])
		}
	}
}

func TestPathErrors(t *testing.T) {
	tests := []struct {
		description   string
		rawPath       string
		expectedError string
	}{
		{
			description:   "invalid command",
			rawPath:       "M 10 20 x",
			expectedError: "Invalid command 'x'",
		},
		{
			description:   "no moveto command at beginnning",
			rawPath:       "10,20",
			expectedError: "Path data does not start with a moveto command: 10,20",
		},
		{
			description:   "incorrect number of parameters",
			rawPath:       "M 10 20 30 Z",
			expectedError: "Incorrect number of parameters for M",
		},
		{
			description:   "parameter not a number",
			rawPath:       "M 10 7%4 Z",
			expectedError: "Unrecognized symbol '%'",
		},
		{
			description:   "parameter not a number",
			rawPath:       "M 10--1 Z",
			expectedError: "Invalid parameter syntax",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			path, err := NewPath(test.rawPath)
			if path != nil {
				t.Fatalf("Path: expected path to be nil, actual: %v", path)
			}

			if err.Error() != test.expectedError {
				t.Fatalf("Path: expected %v, actual %v", test.expectedError, err)
			}
		})
	}
}

func TestPathSubpaths(t *testing.T) {
	tests := []struct {
		description      string
		rawPath          string
		expectedSubpaths []*Path
	}{
		{
			description: "implicit start command",
			rawPath:     "M 10,20 30,40 Z L 35,45 Z",
			expectedSubpaths: []*Path{
				{
					Commands: []*PathCommand{
						{Symbol: "M", Params: []float64{10, 20}},
						{Symbol: "L", Params: []float64{30, 40}},
					},
				},
				{
					Commands: []*PathCommand{
						{Symbol: "M", Params: []float64{10, 20}},
						{Symbol: "L", Params: []float64{35, 45}},
					},
				},
			},
		},
		{
			description: "implicit end command",
			rawPath:     "M 10,20 M 35,45 55,65 Z",
			expectedSubpaths: []*Path{
				{
					Commands: []*PathCommand{
						{Symbol: "M", Params: []float64{10, 20}},
					},
				},
				{
					Commands: []*PathCommand{
						{Symbol: "M", Params: []float64{35, 45}},
						{Symbol: "L", Params: []float64{55, 65}},
					},
				},
			},
		},
		{
			description: "no end command at the end of the path",
			rawPath:     "M 10,20 30,40",
			expectedSubpaths: []*Path{
				{
					Commands: []*PathCommand{
						{Symbol: "M", Params: []float64{10, 20}},
						{Symbol: "L", Params: []float64{30, 40}},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			path, err := NewPath(test.rawPath)
			if err != nil {
				t.Fatalf("Path: unexpected error: %v", err)
			}

			subpaths := path.Subpaths()
			for i, subpath := range test.expectedSubpaths {
				if !subpath.Equal(subpaths[i]) {
					fmt.Println(subpaths[i].Commands[0])
					t.Errorf("Path: expected %v, actual %v", subpath, subpaths[i])
				}
			}
		})
	}
}

func TestPathCommandIsAbsolute(t *testing.T) {
	tests := []struct {
		description string
		command     PathCommand
		expected    bool
	}{
		{
			description: "absolute command",
			command:     PathCommand{Symbol: "L"},
			expected:    true,
		},
		{
			description: "not absolute command",
			command:     PathCommand{Symbol: "l"},
			expected:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			absolute := test.command.IsAbsolute()
			if test.expected != absolute {
				t.Errorf("Path: expected %v, actual %v", test.expected, absolute)
			}
		})
	}
}
