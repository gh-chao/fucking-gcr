package main

import (
	"regexp"
)

var (
	// alphaNumeric defines the alpha numeric atom, typically a
	// component of names. This only allows lower case characters and digits.
	alphaNumeric = `[a-z0-9]+`

	// separator defines the separators allowed to be embedded in name
	// components. This allow one period, one or two underscore and multiple
	// dashes. Repeated dashes and underscores are intentionally treated
	// differently. In order to support valid hostnames as name components,
	// supporting repeated dash was added. Additionally double underscore is
	// now allowed as a separator to loosen the restriction for previously
	// supported names.
	separator = `(?:[._]|__|[-]*)`

	// nameComponent restricts registry path component names to start
	// with at least one letter or number, with following parts able to be
	// separated by one period, one or two underscore and multiple dashes.
	nameComponent = expression(
		alphaNumeric,
		optional(repeated(separator, alphaNumeric)))

	// domainComponent restricts the registry domain component of a
	// repository name to start with a component as defined by DomainRegexp
	// and followed by an optional port.
	domainComponent = `(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])`

	domain = expression(
		domainComponent,
		optional(repeated(literal(`.`), domainComponent)),
		optional(literal(`:`), `[0-9]+`))
	// DomainRegexp defines the structure of potential domain components
	// that may be part of image names. This is purposely a subset of what is
	// allowed by DNS to ensure backwards compatibility with Docker image
	// names.

	tag = `[\w][\w.-]{0,127}`
	// TagRegexp matches valid tag names. From docker/docker:graph/tags.go.

	// anchoredTagRegexp matches valid tag names, anchored at the start and
	// end of the matched string.

	digestPat = `[A-Za-z][A-Za-z0-9]*(?:[-_+.][A-Za-z][A-Za-z0-9]*)*[:][[:xdigit:]]{32,}`
	// DigestRegexp matches valid digests.

	namePat = expression(
		optional(domain, literal(`/`)),
		nameComponent,
		optional(repeated(literal(`/`), nameComponent)))

	ReferenceRegexp = expression(capture(namePat),
		optional(literal(":"), capture(tag)),
		optional(literal("@"), capture(digestPat)))
)

// literal compiles s into a literal regular expression, escaping any regexp
// reserved characters.
func literal(s string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(s))

	if _, complete := re.LiteralPrefix(); !complete {
		panic("must be a literal")
	}

	return re.String()
}

// expression defines a full expression, where each regular expression must
// follow the previous.
func expression(res ...string) string {
	var s string
	for _, re := range res {
		s += re
	}

	return s
}

// optional wraps the expression in a non-capturing group and makes the
// production optional.
func optional(res ...string) string {
	return group(expression(res...)) + `?`
}

// repeated wraps the regexp in a non-capturing group to get one or more
// matches.
func repeated(res ...string) string {
	return group(expression(res...)) + `+`
}

// group wraps the regexp in a non-capturing group.
func group(res ...string) string {
	return `(?:` + expression(res...) + `)`
}

// capture wraps the expression in a capturing group.
func capture(res ...string) string {
	return `(` + expression(res...) + `)`
}
