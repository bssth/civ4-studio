package editor

import (
	"bytes"
	"io"
)

// xmlCommentFixer passes XML through and replaces dashes inside comments with spaces. Mods often have
// comments like "<!-- ---- -->" or "<!--- text --->": the game reads them, but encoding/xml rejects
// "--" inside a comment even in non-strict mode. Bytes are replaced one for one, so line numbers
// in parse errors stay correct. CDATA sections are passed as they are.
type xmlCommentFixer struct {
	r     io.Reader
	err   error // error from r, returned once all data is passed
	state int
	buf   []byte
	start int // buf[start:done] is checked and not returned yet
	done  int
	end   int // buf[done:end] is read from r but not checked yet
}

const (
	xmlFixerText = iota
	xmlFixerComment
	xmlFixerCData
)

// The longest sequence checked at once is "<![CDATA["
const xmlFixerLookahead = len("<![CDATA[")

func newXMLCommentFixer(r io.Reader) *xmlCommentFixer {
	return &xmlCommentFixer{r: r, buf: make([]byte, 32*1024)}
}

func (c *xmlCommentFixer) Read(p []byte) (int, error) {
	for c.start == c.done {
		if c.err != nil && c.done == c.end {
			return 0, c.err
		}
		if c.err == nil {
			c.fill()
		}
		c.check()
	}

	n := copy(p, c.buf[c.start:c.done])
	c.start += n
	return n, nil
}

// fill reads more data, moving the bytes not returned yet to the beginning of the buffer
func (c *xmlCommentFixer) fill() {
	if c.start > 0 {
		copy(c.buf, c.buf[c.start:c.end])
		c.done -= c.start
		c.end -= c.start
		c.start = 0
	}
	n, err := c.r.Read(c.buf[c.end:])
	c.end += n
	if err != nil {
		c.err = err
	}
}

// check moves done forward. A sequence that may be cut by the end of the buffer is left
// for the next read, unless r has no more data.
func (c *xmlCommentFixer) check() {
	for c.done < c.end {
		rest := c.buf[c.done:c.end]
		var mark byte
		switch c.state {
		case xmlFixerText:
			mark = '<'
		case xmlFixerComment:
			mark = '-'
		default:
			mark = ']'
		}

		i := bytes.IndexByte(rest, mark)
		if i < 0 {
			c.done = c.end
			return
		}
		rest = rest[i:]
		if len(rest) < xmlFixerLookahead && c.err == nil {
			c.done += i
			return
		}

		skip := 1
		switch c.state {
		case xmlFixerText:
			if bytes.HasPrefix(rest, []byte("<!--")) {
				c.state, skip = xmlFixerComment, len("<!--")
			} else if bytes.HasPrefix(rest, []byte("<![CDATA[")) {
				c.state, skip = xmlFixerCData, len("<![CDATA[")
			}
		case xmlFixerComment:
			if bytes.HasPrefix(rest, []byte("-->")) {
				c.state, skip = xmlFixerText, len("-->")
			} else {
				rest[0] = ' '
			}
		default:
			if bytes.HasPrefix(rest, []byte("]]>")) {
				c.state, skip = xmlFixerText, len("]]>")
			}
		}
		c.done += i + skip
	}
}
