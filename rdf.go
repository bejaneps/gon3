package gon3

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Term interface {
	String() string
	RawValue() string
	Equal(Term) bool
}

// This must be a full (i.e. not relative IRI)
type IRI struct {
	url *url.URL
}

func (i IRI) String() string {
	return fmt.Sprintf("<%s>", i.url)
}

func (i IRI) Equals(other IRI) bool {
	return i.String() == other.String()
}

func newIRIFromString(s string) (IRI, error) {
	url, err := iriRefToURL(s)
	return IRI{url}, err
}

func iriRefToURL(s string) (*url.URL, error) {
	// strip <>, unescape, parse into url
	if strings.HasPrefix(s, "<") {
		s = s[1 : len(s)-1]
	}
	unescaped := unescapeUChar(s)
	return url.Parse(unescaped)
}

// see http://www.w3.org/TR/rdf11-concepts/#dfn-blank-node
type BlankNode struct {
	Id    int
	Label string
}

func (b BlankNode) String() string {
	return fmt.Sprintf("_:%s", b.Label)
}

// see http://www.w3.org/TR/rdf11-concepts/#dfn-literal
type Literal struct {
	LexicalForm string
	DatatypeIRI IRI
	LanguageTag string
}

func (l Literal) String() string {
	if l.LanguageTag != "" {
		return fmt.Sprintf("%q@%s", l.LexicalForm, l.LanguageTag)
	}
	return fmt.Sprintf("%q^^%s", l.LexicalForm, l.DatatypeIRI)
}

func (l Literal) Equals(other Literal) bool {
	return l.LexicalForm == other.LexicalForm && l.DatatypeIRI.Equals(other.DatatypeIRI) && l.LanguageTag == other.LanguageTag
}

func lexicalForm(s string) string {
	var unquoted string
	if strings.HasPrefix(s, `"""`) || strings.HasPrefix(s, `'''`) {
		unquoted = s[3 : len(s)-3]
	} else {
		unquoted = s[1 : len(s)-1]
	}
	// TODO: resolve escapes
	u := unescapeUChar(unquoted)
	ret := unescapeEChar(u)
	return ret
}

func unescapeEChar(s string) string {
	var replacements = []struct {
		old string
		new string
	}{
		{`\t`, "\t"},
		{`\b`, "\b"},
		{`\n`, "\n"},
		{`\r`, "\r"},
		{`\f`, "\f"},
		{`\"`, `"`},
		{`\'`, `'`},
		{`\\`, `\`},
	}
	for _, r := range replacements {
		s = strings.Replace(s, r.old, r.new, -1)
	}
	return s
}

func unescapeUChar(s string) string {
	for {
		var start, hex, end string
		uIdx := strings.Index(s, `\u`)
		UIdx := strings.Index(s, `\U`)
		if uIdx >= 0 {
			start = s[:uIdx]
			if uIdx+6 > len(s) {
				hex = s[uIdx+2:]
				end = ""
			} else {
				hex = s[uIdx+2 : uIdx+6]
				end = s[uIdx+6:]
			}
		} else if UIdx >= 0 {
			start = s[:UIdx]
			if UIdx+10 > len(s) {
				hex = s[UIdx+2:]
				end = ""
			} else {
				hex = s[UIdx+2 : UIdx+10]
				end = s[UIdx+10:]
			}
		} else {
			break
		}
		num, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			panic(err) // TODO: this shouldn't happen
		}
		s = fmt.Sprintf("%s%s%s", start, string(rune(num)), end)
	}
	return s
}

// see http://www.w3.org/TR/rdf11-concepts/#dfn-rdf-triple
type Triple struct {
	Subject   RDFTerm // cannot be a literal
	Predicate IRI
	Object    RDFTerm
}

func (t *Triple) String() string {
	return fmt.Sprintf("%s %s %s .", t.Subject, t.Predicate, t.Object)
}

// An RDF graph is a set of RDF triples
type Graph []*Triple

func (g Graph) String() string {
	str := ""
	for i, t := range g {
		if i > 0 {
			str += "\n"
		}
		str = fmt.Sprintf("%s%s", str, t)
	}
	return str
}
