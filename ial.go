// IAL implements

package mmark

import (
	"bytes"
)

// One or more of these can be attached to block elements

type IAL struct {
	id    string            // #id
	class []string          // 0 or more .class
	attr  map[string]string // key=value pairs
}

// Parsing and thus detecting an IAL. Return a valid *IAL or nil.
// IAL can have #id, .class or key=value element seperated by spaces, that may be escaped
func (p *parser) isIAL(data []byte) int {
	esc := false
	quote := false
	ialB := 0
	ial := &IAL{attr: make(map[string]string)}
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case ' ':
			if quote {
				continue
			}
			chunk := data[ialB+1 : i]
			switch {
			case chunk[0] == '.':
				ial.class = append(ial.class, string(chunk[1:]))
			case chunk[0] == '#':
				ial.id = string(chunk[1:])
			default:
				k, v := parseKeyValue(chunk)
				if k != "" {
					ial.attr[k] = v
				}
			}
			ialB = i
		case '"':
			if esc {
				esc = !esc
				continue
			}
			quote = !quote
		case '\\':
			esc = !esc
		case '}':
			if esc {
				esc = !esc
				continue
			}
			// if this is mainmatter, frontmatter, or backmatter it isn't an IAL.
			s := string(data[1:i])
			switch s {
			case "frontmatter":
				fallthrough
			case "mainmatter":
				fallthrough
			case "backmatter":
				return 0
			}
			chunk := data[ialB+1 : i]
			switch {
			case chunk[0] == '.':
				ial.class = append(ial.class, string(chunk[1:]))
			case chunk[0] == '#':
				ial.id = string(chunk[1:])
			default:
				k, v := parseKeyValue(chunk)
				if k != "" {
					ial.attr[k] = v
				}
			}
			p.ial = append(p.ial, ial)
			return i + 1
		default:
			esc = false
		}
	}
	return 0
}

// renderIAL renders an IAL and returns a string that can be included in the tag:
// class="class" anchor="id" key="value"
func renderIAL(i []*IAL) string {
	anchor := ""
	class := ""
	attr := ""
	for _, i1 := range i {
		if i1.id != "" {
			anchor = "anchor=\"" + i1.id + "\""
		}
		for _, c1 := range i1.class {
			if class == "" {
				class += c1
				continue
			}
			class += " " + c1
		}
		for k, v := range i1.attr {
			if attr == "" {
				attr = k + "=\"" + v + "\""
				continue
			}
			attr += " " + k + "=\"" + v + "\""

		}
	}
	s := ""
	if anchor != "" {
		s = " " + anchor
	}
	if class != "" {
		s += " class=\"" + class + "\""
	}
	if attr != "" {
		s += " " + attr
	}
	return s
}

func parseKeyValue(chunk []byte) (string, string) {
	chunks := bytes.SplitN(chunk, []byte{'='}, 2)
	if len(chunks) != 2 {
		return "", ""
	}
	chunks[1] = bytes.Replace(chunks[1], []byte{'"'}, nil, -1)
	return string(chunks[0]), string(chunks[1])
}
