package label

import (
	"bytes"
	"fmt"
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"
	"sort"
	"strconv"
	"strings"
)

type Requirements []Requirement

type Selector interface {
	Matches(Labels) bool

	Empty() bool

	String() string

	Add(r ...Requirement) Selector

	Requirements() (requirements Requirements, selectable bool)

	DeepCopySelector() Selector

	RequiresExactMatch(label string) (value string, found bool)
}

func Everything() Selector {
	return internalSelector{}
}

type nothingSelector struct{}

func (n nothingSelector) Matches(_ Labels) bool              { return false }
func (n nothingSelector) Empty() bool                        { return false }
func (n nothingSelector) String() string                     { return "" }
func (n nothingSelector) Add(_ ...Requirement) Selector      { return n }
func (n nothingSelector) Requirements() (Requirements, bool) { return nil, false }
func (n nothingSelector) DeepCopySelector() Selector         { return n }
func (n nothingSelector) RequiresExactMatch(label string) (value string, found bool) {
	return "", false
}

func Nothing() Selector {
	return nothingSelector{}
}

func NewSelector() Selector {
	return internalSelector(nil)
}

type internalSelector []Requirement

func (s internalSelector) DeepCopy() internalSelector {
	if s == nil {
		return nil
	}
	result := make([]Requirement, len(s))
	for i := range s {
		s[i].DeepCopyInto(&result[i])
	}
	return result
}

func (s internalSelector) DeepCopySelector() Selector {
	return s.DeepCopy()
}

type ByKey []Requirement

func (a ByKey) Len() int { return len(a) }

func (a ByKey) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a ByKey) Less(i, j int) bool { return a[i].key < a[j].key }

type Requirement struct {
	key       string
	operator  Operator
	strValues []string
}

func NewRequirement(key string, op Operator, vals []string) (*Requirement, error) {
	if err := validateLabelKey(key); err != nil {
		return nil, err
	}
	switch op {
	case In, NotIn:
		if len(vals) == 0 {
			return nil, fmt.Errorf("for 'in', 'notin' operators, values set can't be empty")
		}
	case Equals, DoubleEquals, NotEquals:
		if len(vals) != 1 {
			return nil, fmt.Errorf("exact-match compatibility requires one single value")
		}
	case Exists, DoesNotExist:
		if len(vals) != 0 {
			return nil, fmt.Errorf("values set must be empty for exists and does not exist")
		}
	case GreaterThan, LessThan:
		if len(vals) != 1 {
			return nil, fmt.Errorf("for 'Gt', 'Lt' operators, exactly one value is required")
		}
		for i := range vals {
			if _, err := strconv.ParseInt(vals[i], 10, 64); err != nil {
				return nil, fmt.Errorf("for 'Gt', 'Lt' operators, the value must be an integer")
			}
		}
	default:
		return nil, fmt.Errorf("operator '%v' is not recognized", op)
	}

	for i := range vals {
		if err := validateLabelValue(key, vals[i]); err != nil {
			return nil, err
		}
	}
	return &Requirement{key: key, operator: op, strValues: vals}, nil
}

func (r *Requirement) hasValue(value string) bool {
	for i := range r.strValues {
		if r.strValues[i] == value {
			return true
		}
	}
	return false
}

func (r *Requirement) Matches(ls Labels) bool {
	switch r.operator {
	case In, Equals, DoubleEquals:
		if !ls.Has(r.key) {
			return false
		}
		return r.hasValue(ls.Get(r.key))
	case NotIn, NotEquals:
		if !ls.Has(r.key) {
			return true
		}
		return !r.hasValue(ls.Get(r.key))
	case Exists:
		return ls.Has(r.key)
	case DoesNotExist:
		return !ls.Has(r.key)
	case GreaterThan, LessThan:
		if !ls.Has(r.key) {
			return false
		}
		lsValue, err := strconv.ParseInt(ls.Get(r.key), 10, 64)
		if err != nil {
			klog.V(10).Infof("ParseInt failed for value %+v in label %+v, %+v", ls.Get(r.key), ls, err)
			return false
		}

		if len(r.strValues) != 1 {
			klog.V(10).Infof("Invalid values count %+v of requirement %#v, for 'Gt', 'Lt' operators, exactly one value is required", len(r.strValues), r)
			return false
		}

		var rValue int64
		for i := range r.strValues {
			rValue, err = strconv.ParseInt(r.strValues[i], 10, 64)
			if err != nil {
				klog.V(10).Infof("ParseInt failed for value %+v in requirement %#v, for 'Gt', 'Lt' operators, the value must be an integer", r.strValues[i], r)
				return false
			}
		}
		return (r.operator == GreaterThan && lsValue > rValue) || (r.operator == LessThan && lsValue < rValue)
	default:
		return false
	}
}

func (r *Requirement) Key() string {
	return r.key
}

func (r *Requirement) Operator() Operator {
	return r.operator
}

func (r *Requirement) Values() sets.String {
	ret := sets.String{}
	for i := range r.strValues {
		ret.Insert(r.strValues[i])
	}
	return ret
}

func (lsel internalSelector) Empty() bool {
	if lsel == nil {
		return true
	}
	return len(lsel) == 0
}

func (in *Requirement) DeepCopyInto(out *Requirement) {
	*out = *in
	if in.strValues != nil {
		in, out := &in.strValues, &out.strValues
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *Requirement) DeepCopy() *Requirement {
	if in == nil {
		return nil
	}
	out := new(Requirement)
	in.DeepCopyInto(out)
	return out
}

func (r *Requirement) String() string {
	var buffer bytes.Buffer
	if r.operator == DoesNotExist {
		buffer.WriteString("!")
	}
	buffer.WriteString(r.key)

	switch r.operator {
	case Equals:
		buffer.WriteString("=")
	case DoubleEquals:
		buffer.WriteString("==")
	case NotEquals:
		buffer.WriteString("!=")
	case In:
		buffer.WriteString(" in ")
	case NotIn:
		buffer.WriteString(" notin ")
	case GreaterThan:
		buffer.WriteString(">")
	case LessThan:
		buffer.WriteString("<")
	case Exists, DoesNotExist:
		return buffer.String()
	}

	switch r.operator {
	case In, NotIn:
		buffer.WriteString("(")
	}
	if len(r.strValues) == 1 {
		buffer.WriteString(r.strValues[0])
	} else { // only > 1 since == 0 prohibited by NewRequirement
		buffer.WriteString(strings.Join(safeSort(r.strValues), ","))
	}

	switch r.operator {
	case In, NotIn:
		buffer.WriteString(")")
	}
	return buffer.String()
}

func safeSort(in []string) []string {
	if sort.StringsAreSorted(in) {
		return in
	}
	out := make([]string, len(in))
	copy(out, in)
	sort.Strings(out)
	return out
}

func (lsel internalSelector) Add(reqs ...Requirement) Selector {
	var sel internalSelector
	for ix := range lsel {
		sel = append(sel, lsel[ix])
	}
	for _, r := range reqs {
		sel = append(sel, r)
	}
	sort.Sort(ByKey(sel))
	return sel
}

func (lsel internalSelector) Matches(l Labels) bool {
	for ix := range lsel {
		if matches := lsel[ix].Matches(l); !matches {
			return false
		}
	}
	return true
}

func (lsel internalSelector) Requirements() (Requirements, bool) { return Requirements(lsel), true }

func (lsel internalSelector) String() string {
	var reqs []string
	for ix := range lsel {
		reqs = append(reqs, lsel[ix].String())
	}
	return strings.Join(reqs, ",")
}

func (lsel internalSelector) RequiresExactMatch(label string) (value string, found bool) {
	for ix := range lsel {
		if lsel[ix].key == label {
			switch lsel[ix].operator {
			case Equals, DoubleEquals, In:
				if len(lsel[ix].strValues) == 1 {
					return lsel[ix].strValues[0], true
				}
			}
			return "", false
		}
	}
	return "", false
}

type Token int

const (
	ErrorToken Token = iota
	EndOfStringToken
	ClosedParToken
	CommaToken
	DoesNotExistToken
	DoubleEqualsToken
	EqualsToken
	GreaterThanToken
	IdentifierToken
	InToken
	LessThanToken
	NotEqualsToken
	NotInToken
	OpenParToken
)

var string2token = map[string]Token{
	")":     ClosedParToken,
	",":     CommaToken,
	"!":     DoesNotExistToken,
	"==":    DoubleEqualsToken,
	"=":     EqualsToken,
	">":     GreaterThanToken,
	"in":    InToken,
	"<":     LessThanToken,
	"!=":    NotEqualsToken,
	"notin": NotInToken,
	"(":     OpenParToken,
}

type ScannedItem struct {
	tok     Token
	literal string
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

func isSpecialSymbol(ch byte) bool {
	switch ch {
	case '=', '!', '(', ')', ',', '>', '<':
		return true
	}
	return false
}

type Lexer struct {
	s   string
	pos int
}

func (l *Lexer) read() (b byte) {
	b = 0
	if l.pos < len(l.s) {
		b = l.s[l.pos]
		l.pos++
	}
	return b
}

func (l *Lexer) unread() {
	l.pos--
}

func (l *Lexer) scanIDOrKeyword() (tok Token, lit string) {
	var buffer []byte
IdentifierLoop:
	for {
		switch ch := l.read(); {
		case ch == 0:
			break IdentifierLoop
		case isSpecialSymbol(ch) || isWhitespace(ch):
			l.unread()
			break IdentifierLoop
		default:
			buffer = append(buffer, ch)
		}
	}
	s := string(buffer)
	if val, ok := string2token[s]; ok { // is a literal token?
		return val, s
	}
	return IdentifierToken, s // otherwise is an identifier
}

func (l *Lexer) scanSpecialSymbol() (Token, string) {
	lastScannedItem := ScannedItem{}
	var buffer []byte
SpecialSymbolLoop:
	for {
		switch ch := l.read(); {
		case ch == 0:
			break SpecialSymbolLoop
		case isSpecialSymbol(ch):
			buffer = append(buffer, ch)
			if token, ok := string2token[string(buffer)]; ok {
				lastScannedItem = ScannedItem{tok: token, literal: string(buffer)}
			} else if lastScannedItem.tok != 0 {
				l.unread()
				break SpecialSymbolLoop
			}
		default:
			l.unread()
			break SpecialSymbolLoop
		}
	}
	if lastScannedItem.tok == 0 {
		return ErrorToken, fmt.Sprintf("error expected: keyword found '%s'", buffer)
	}
	return lastScannedItem.tok, lastScannedItem.literal
}

func (l *Lexer) skipWhiteSpaces(ch byte) byte {
	for {
		if !isWhitespace(ch) {
			return ch
		}
		ch = l.read()
	}
}

func (l *Lexer) Lex() (tok Token, lit string) {
	switch ch := l.skipWhiteSpaces(l.read()); {
	case ch == 0:
		return EndOfStringToken, ""
	case isSpecialSymbol(ch):
		l.unread()
		return l.scanSpecialSymbol()
	default:
		l.unread()
		return l.scanIDOrKeyword()
	}
}

type Parser struct {
	l            *Lexer
	scannedItems []ScannedItem
	position     int
}

type ParserContext int

const (
	KeyAndOperator ParserContext = iota
	Values
)

func (p *Parser) lookahead(context ParserContext) (Token, string) {
	tok, lit := p.scannedItems[p.position].tok, p.scannedItems[p.position].literal
	if context == Values {
		switch tok {
		case InToken, NotInToken:
			tok = IdentifierToken
		}
	}
	return tok, lit
}

func (p *Parser) consume(context ParserContext) (Token, string) {
	p.position++
	tok, lit := p.scannedItems[p.position-1].tok, p.scannedItems[p.position-1].literal
	if context == Values {
		switch tok {
		case InToken, NotInToken:
			tok = IdentifierToken
		}
	}
	return tok, lit
}

func (p *Parser) scan() {
	for {
		token, literal := p.l.Lex()
		p.scannedItems = append(p.scannedItems, ScannedItem{token, literal})
		if token == EndOfStringToken {
			break
		}
	}
}

func (p *Parser) parse() (internalSelector, error) {
	p.scan() // init scannedItems

	var requirements internalSelector
	for {
		tok, lit := p.lookahead(Values)
		switch tok {
		case IdentifierToken, DoesNotExistToken:
			r, err := p.parseRequirement()
			if err != nil {
				return nil, fmt.Errorf("unable to parse requirement: %v", err)
			}
			requirements = append(requirements, *r)
			t, l := p.consume(Values)
			switch t {
			case EndOfStringToken:
				return requirements, nil
			case CommaToken:
				t2, l2 := p.lookahead(Values)
				if t2 != IdentifierToken && t2 != DoesNotExistToken {
					return nil, fmt.Errorf("found '%s', expected: identifier after ','", l2)
				}
			default:
				return nil, fmt.Errorf("found '%s', expected: ',' or 'end of string'", l)
			}
		case EndOfStringToken:
			return requirements, nil
		default:
			return nil, fmt.Errorf("found '%s', expected: !, identifier, or 'end of string'", lit)
		}
	}
}

func (p *Parser) parseRequirement() (*Requirement, error) {
	key, operator, err := p.parseKeyAndInferOperator()
	if err != nil {
		return nil, err
	}
	if operator == Exists || operator == DoesNotExist { // operator found lookahead set checked
		return NewRequirement(key, operator, []string{})
	}
	operator, err = p.parseOperator()
	if err != nil {
		return nil, err
	}
	var values sets.String
	switch operator {
	case In, NotIn:
		values, err = p.parseValues()
	case Equals, DoubleEquals, NotEquals, GreaterThan, LessThan:
		values, err = p.parseExactValue()
	}
	if err != nil {
		return nil, err
	}
	return NewRequirement(key, operator, values.List())

}

func (p *Parser) parseKeyAndInferOperator() (string, Operator, error) {
	var operator Operator
	tok, literal := p.consume(Values)
	if tok == DoesNotExistToken {
		operator = DoesNotExist
		tok, literal = p.consume(Values)
	}
	if tok != IdentifierToken {
		err := fmt.Errorf("found '%s', expected: identifier", literal)
		return "", "", err
	}
	if err := validateLabelKey(literal); err != nil {
		return "", "", err
	}
	if t, _ := p.lookahead(Values); t == EndOfStringToken || t == CommaToken {
		if operator != DoesNotExist {
			operator = Exists
		}
	}
	return literal, operator, nil
}

func (p *Parser) parseOperator() (op Operator, err error) {
	tok, lit := p.consume(KeyAndOperator)
	switch tok {
	case InToken:
		op = In
	case EqualsToken:
		op = Equals
	case DoubleEqualsToken:
		op = DoubleEquals
	case GreaterThanToken:
		op = GreaterThan
	case LessThanToken:
		op = LessThan
	case NotInToken:
		op = NotIn
	case NotEqualsToken:
		op = NotEquals
	default:
		return "", fmt.Errorf("found '%s', expected: '=', '!=', '==', 'in', notin'", lit)
	}
	return op, nil
}

func (p *Parser) parseValues() (sets.String, error) {
	tok, lit := p.consume(Values)
	if tok != OpenParToken {
		return nil, fmt.Errorf("found '%s' expected: '('", lit)
	}
	tok, lit = p.lookahead(Values)
	switch tok {
	case IdentifierToken, CommaToken:
		s, err := p.parseIdentifiersList() // handles general cases
		if err != nil {
			return s, err
		}
		if tok, _ = p.consume(Values); tok != ClosedParToken {
			return nil, fmt.Errorf("found '%s', expected: ')'", lit)
		}
		return s, nil
	case ClosedParToken: // handles "()"
		p.consume(Values)
		return sets.NewString(""), nil
	default:
		return nil, fmt.Errorf("found '%s', expected: ',', ')' or identifier", lit)
	}
}

func (p *Parser) parseIdentifiersList() (sets.String, error) {
	s := sets.NewString()
	for {
		tok, lit := p.consume(Values)
		switch tok {
		case IdentifierToken:
			s.Insert(lit)
			tok2, lit2 := p.lookahead(Values)
			switch tok2 {
			case CommaToken:
				continue
			case ClosedParToken:
				return s, nil
			default:
				return nil, fmt.Errorf("found '%s', expected: ',' or ')'", lit2)
			}
		case CommaToken: // handled here since we can have "(,"
			if s.Len() == 0 {
				s.Insert("") // to handle (,
			}
			tok2, _ := p.lookahead(Values)
			if tok2 == ClosedParToken {
				s.Insert("") // to handle ,)  Double "" removed by StringSet
				return s, nil
			}
			if tok2 == CommaToken {
				p.consume(Values)
				s.Insert("") // to handle ,, Double "" removed by StringSet
			}
		default: // it can be operator
			return s, fmt.Errorf("found '%s', expected: ',', or identifier", lit)
		}
	}
}

func (p *Parser) parseExactValue() (sets.String, error) {
	s := sets.NewString()
	tok, lit := p.lookahead(Values)
	if tok == EndOfStringToken || tok == CommaToken {
		s.Insert("")
		return s, nil
	}
	tok, lit = p.consume(Values)
	if tok == IdentifierToken {
		s.Insert(lit)
		return s, nil
	}
	return nil, fmt.Errorf("found '%s', expected: identifier", lit)
}

func Parse(selector string) (Selector, error) {
	parsedSelector, err := parse(selector)
	if err == nil {
		return parsedSelector, nil
	}
	return nil, err
}

func parse(selector string) (internalSelector, error) {
	p := &Parser{l: &Lexer{s: selector, pos: 0}}
	items, err := p.parse()
	if err != nil {
		return nil, err
	}
	sort.Sort(ByKey(items)) // sort to grant determistic parsing
	return internalSelector(items), err
}

func validateLabelKey(k string) error {
	if errs := validation.IsQualifiedName(k); len(errs) != 0 {
		return fmt.Errorf("invalid label key %q: %s", k, strings.Join(errs, "; "))
	}
	return nil
}

func validateLabelValue(k, v string) error {
	if errs := validation.IsValidLabelValue(v); len(errs) != 0 {
		return fmt.Errorf("invalid label value: %q: at key: %q: %s", v, k, strings.Join(errs, "; "))
	}
	return nil
}

func SelectorFromSet(ls Set) Selector {
	return SelectorFromValidatedSet(ls)
}

func ValidatedSelectorFromSet(ls Set) (Selector, error) {
	if ls == nil || len(ls) == 0 {
		return internalSelector{}, nil
	}
	requirements := make([]Requirement, 0, len(ls))
	for label, value := range ls {
		r, err := NewRequirement(label, Equals, []string{value})
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, *r)
	}
	sort.Sort(ByKey(requirements))
	return internalSelector(requirements), nil
}

func SelectorFromValidatedSet(ls Set) Selector {
	if ls == nil || len(ls) == 0 {
		return internalSelector{}
	}
	requirements := make([]Requirement, 0, len(ls))
	for label, value := range ls {
		requirements = append(requirements, Requirement{key: label, operator: Equals, strValues: []string{value}})
	}
	sort.Sort(ByKey(requirements))
	return internalSelector(requirements)
}

func ParseToRequirements(selector string) ([]Requirement, error) {
	return parse(selector)
}
