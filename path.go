package svg

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// PathCommand is a representation of an SVG path command. It contains the
// operator symbol and the command's parameters.
type PathCommand struct {
	Symbol string
	Params []float64
}

// IsAbsolute returns true is the SVG path command is absolute.
func (c *PathCommand) IsAbsolute() bool {
	return c.Symbol == strings.ToUpper(c.Symbol)
}

// Equal compares two commands.
func (c *PathCommand) Equal(o *PathCommand) bool {
	if c.Symbol != o.Symbol {
		return false
	}

	for i, param := range c.Params {
		if param != o.Params[i] {
			return false
		}
	}

	return true
}

// Path is a collection of all the subpaths in path data attribute.
type Path struct {
	Commands []*PathCommand
}

// Equal compares two paths.
func (p *Path) Equal(o *Path) bool {
	if len(p.Commands) != len(o.Commands) {
		return false
	}

	for i, command := range p.Commands {
		if !command.Equal(o.Commands[i]) {
			return false
		}
	}

	return true
}

// NewPath takes value of a path data attribute transforms it into a series of
// commands containing the appropriate parameters.
func NewPath(raw string) (*Path, error) {
	cmds, err := commands(raw)
	if err != nil {
		return nil, err
	}

	return &Path{Commands: cmds}, nil
}

// Subpaths computes all subpaths from a given path. 'Z' command is excluded from
// the resulting paths.
func (p *Path) Subpaths() []*Path {
	path := &Path{}
	var subpaths []*Path
	var mostRecentStart *PathCommand

	for _, command := range p.Commands {
		switch strings.ToLower(command.Symbol) {
		case startCommand:
			if len(path.Commands) > 0 {
				subpaths = append(subpaths, path)
			}
			path = &Path{Commands: []*PathCommand{command}}
			mostRecentStart = &PathCommand{
				Symbol: command.Symbol,
				Params: command.Params,
			}
		case endCommand:
			subpaths = append(subpaths, path)
			path = &Path{}
		default:
			if len(path.Commands) == 0 {
				path = &Path{Commands: []*PathCommand{mostRecentStart}}
			}
			path.Commands = append(path.Commands, command)
		}
	}

	if len(path.Commands) > 0 {
		subpaths = append(subpaths, path)
	}

	return subpaths
}

const (
	startCommand = "m"
	endCommand   = "z"
)

// commandParams maps a command symbol to the number of parameters that
// command requires.
var commandParams = map[string]int{
	"m": 2, "z": 0, "l": 2, "h": 1, "v": 1,
	"c": 6, "s": 4, "q": 4, "t": 2, "a": 7,
}

// commands makes a slice of path commands from a raw path data attribute.
func commands(raw string) ([]*PathCommand, error) {
	ts, err := tokenize(raw)
	if err != nil {
		return nil, err
	}

	tokens := *ts

	// From specification, a path data attribute is invalid if it does not
	// start with moveto command.
	if len(tokens) > 0 && strings.ToLower(tokens[0].value) != startCommand {
		return nil, fmt.Errorf(
			"Path data does not start with a moveto command: %s", raw)
	}

	operands := []float64{}
	cmds := []*PathCommand{}

	for i := len(tokens) - 1; i >= 0; i-- {
		value := tokens[i].value
		if !tokens[i].operator {
			number, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("Invalid parameter syntax")
			}
			operands = append(operands, number)
			continue
		}

		paramCount, ok := commandParams[strings.ToLower(value)]
		if !ok {
			return nil, fmt.Errorf("Invalid command '%s'", value)
		}

		operandCount := len(operands)
		if paramCount == 0 && operandCount == 0 {
			command := &PathCommand{Symbol: value}
			cmds = append([]*PathCommand{command}, cmds...)
			continue
		}

		if paramCount == 0 || operandCount%paramCount != 0 {
			return nil, fmt.Errorf("Incorrect number of parameters for %v", value)
		}

		loopCount := operandCount / paramCount
		for i := 0; i < loopCount; i++ {
			operator := value
			if operator == "m" && i < loopCount-1 {
				operator = "l"
			}
			if operator == "M" && i < loopCount-1 {
				operator = "L"
			}
			command := &PathCommand{operator, reverse(operands[:paramCount])}
			cmds = append([]*PathCommand{command}, cmds...)
			operands = operands[paramCount:]
		}
	}

	return cmds, nil
}

// token can contain an operator or an operand as string.
type token struct {
	value    string
	operator bool
}

// tokens is a collection of tokens
type tokens []token

// add appends a token if the value is non-empty.
// Returns true if a new token has been added.
func (ts *tokens) add(value []rune, operator bool) bool {
	if len(value) == 0 {
		return false
	}

	*ts = append(*ts, token{string(value), operator})

	return true
}

// tokenize takes value of path data attribute and transforms it into a slice of
// tokens than represent operators and operands.
func tokenize(raw string) (*tokens, error) {
	ts := &tokens{}

	var operand []rune
	for _, r := range raw {
		switch {
		case r == '.':
			if len(operand) == 0 {
				operand = append(operand, '0')
			}
			if contains(operand, '.') {
				ts.add(operand, false)
				operand = []rune{'0'}
			}
			fallthrough

		case r >= '0' && r <= '9' || r == 'e':
			operand = append(operand, r)

		case r == '-':
			if len(operand) > 0 && operand[len(operand)-1] == 'e' {
				operand = append(operand, r)
				continue
			}
			ts.add(operand, false)
			operand = []rune{r}

		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			if ok := ts.add(operand, false); ok {
				operand = []rune{}
			}
			ts.add([]rune{r}, true)
			continue

		case unicode.IsSpace(r) || r == ',':
			if ok := ts.add(operand, false); ok {
				operand = []rune{}
			}

		default:
			return nil, fmt.Errorf("Unrecognized symbol '%s'", string(r))
		}
	}

	ts.add(operand, false)

	return ts, nil
}

func reverse(ops []float64) []float64 {
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}
	return ops
}

func contains(rs []rune, val rune) bool {
	for _, r := range rs {
		if r == val {
			return true
		}
	}
	return false
}
