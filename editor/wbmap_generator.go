package editor

import (
	"bytes"
	"fmt"
	"strconv"
)

const defaultLineIndent = "\t"

// SimpleGenerator is a simple WbMap format generator. It helps to create a WbMap sections and key-value pairs.
// Subsections are also supported, just call StartSection and EndSection methods again.
type SimpleGenerator struct {
	buffer bytes.Buffer
	indent int
	endTag []string
}

// Bytes returns the generated WbMap file as a byte slice.
// You can use this method and save output to .CivBeyondSwordWBSave file
func (g *SimpleGenerator) Bytes() []byte {
	return g.buffer.Bytes()
}

// StartSection starts a new section with a given startTag and endTag.
// It also increases the indent level and saves the endTag for later use, so subsections supported too.
func (g *SimpleGenerator) StartSection(startTag string, endTag string) {
	for i := 0; i < g.indent; i++ {
		g.buffer.WriteString(defaultLineIndent)
	}

	g.buffer.WriteString(startTag + "\n")
	g.endTag = append(g.endTag, endTag)
	g.indent++
}

// EndSection ends the current section. It decreases the indent level and writes the end tag.
func (g *SimpleGenerator) EndSection() {
	if len(g.endTag) < 1 {
		return
	}

	g.indent--
	for i := 0; i < g.indent; i++ {
		g.buffer.WriteString(defaultLineIndent)
	}

	g.buffer.WriteString(g.endTag[len(g.endTag)-1] + "\n")
	g.endTag = g.endTag[:len(g.endTag)-1]
}

// AddLine adds a line to the current section. It automatically adds the current indent level.
func (g *SimpleGenerator) AddLine(line string) {
	for i := 0; i < g.indent; i++ {
		g.buffer.WriteString(defaultLineIndent)
	}
	g.buffer.WriteString(line)
	g.buffer.WriteString("\n")
}

// AddComment adds a comment to next line without spacing. In doesn't use the current indent level
func (g *SimpleGenerator) AddComment(comment string) {
	g.buffer.WriteString("#" + comment)
}

// AddKeyValue adds a key-value pair to the current section.
func (g *SimpleGenerator) AddKeyValue(key string, value interface{}) {
	if value == nil || value == "" {
		return
	}

	g.AddLine(fmt.Sprintf("%s=%v", key, value))
}

func (g *SimpleGenerator) AddKeyValueInt(key string, value int) {
	g.AddKeyValue(key, strconv.Itoa(value))
}

func (g *SimpleGenerator) AddKeyValueInt64(key string, value int64) {
	g.AddKeyValue(key, strconv.FormatInt(value, 10))
}

func (g *SimpleGenerator) AddKeyValueUint(key string, value uint64) {
	g.AddKeyValue(key, strconv.Itoa(int(value)))
}

func (g *SimpleGenerator) AddKeyValueBool(key string, value bool) {
	g.AddKeyValue(key, BoolToInt(value))
}

func (g *SimpleGenerator) AddKeyValueString(key string, value string) {
	g.AddKeyValue(key, value)
}

func (g *SimpleGenerator) AddKeyValueArray(key string, values []string) {
	for _, value := range values {
		g.AddKeyValue(key, value)
	}
}

func (g *SimpleGenerator) AddKeyValueIntArray(key string, values []int) {
	for _, value := range values {
		g.AddKeyValue(key, strconv.Itoa(value))
	}
}

func (g *SimpleGenerator) AddKeyValueUintArray(key string, values []uint) {
	for _, value := range values {
		g.AddKeyValue(key, strconv.Itoa(int(value)))
	}
}

func (g *SimpleGenerator) AddCommaSeparatedValues(values ...interface{}) {
	for i := 0; i < g.indent; i++ {
		g.buffer.WriteString(defaultLineIndent)
	}

	for i, value := range values {
		if i > 0 {
			g.buffer.WriteString(",")
		}
		g.buffer.WriteString(fmt.Sprintf("%v", value))
	}

	g.buffer.WriteString("\n")
}
