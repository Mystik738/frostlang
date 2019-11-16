package frostlang

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//ConvertLangToJSON converts all lang files in a directory to json
func ConvertLangToJSON(dir string, overwrite bool) {
	if dir[len(dir)-1:] != "/" {
		dir = dir + "/"
	}

	var files []string
	if _, err := os.Stat(dir); err == nil {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".lang" {
				return nil
			}

			files = append(files, info.Name())
			return nil
		})
		for _, filename := range files {
			if _, err := os.Stat(dir + filename[:strings.Index(filename, ".")] + ".json"); overwrite || err != nil {
				content, err := ioutil.ReadFile(dir + filename)
				content = content[8:]
				checkError(err)

				//Initialize our data structure.
				data := NewNode("")

				//File is as follows: 8 unknown bytes followed by entries of the following, all little endian:
				// 2 bytes for tag length
				// Tag of length {2 byte val}
				// 2 bytes for text length
				// Text of length {2 byte val}*2. Text is encoded in 16-bit values, hence the double byte
				i := 0
				for i < len(content) {
					tagLength := binary.LittleEndian.Uint16(content[i : i+2])
					i += 2
					tagString := content[i : i+int(tagLength)]
					//Tags are split by '/'
					tags := bytes.Split(tagString, []byte{byte(47)})
					i += int(tagLength)
					textLength := binary.LittleEndian.Uint16(content[i : i+2])
					i += 2
					splittext := content[i : i+int(textLength*2)]
					i += int(textLength * 2)

					//Once we get all our text, we need to decode each 16-bit value
					text := ""
					for c := 0; c < len(splittext); c += 2 {
						val := binary.LittleEndian.Uint16(splittext[c:(c + 2)])
						text += string(val)
					}

					//We have our tags and text, add to our data structure.
					data.Add(tags, text)
				}

				file, err := os.Create(dir + filename[:strings.Index(filename, ".")] + ".json")
				checkError(err)

				file.WriteString(data.ToJSON())
				err = file.Close()
				checkError(err)
			}
		}
	}
}

//Node are strucutres that should have Text or Children, but not both
//This allows us to store a string-based JSON object of indeterminate depth.
type Node struct {
	Tag      string
	Text     string
	Children map[string]*Node
	Elements int
	Depth    int
}

//AppendChild Simple way to add a child by the child's tag
func (n *Node) AppendChild(tag string) Node {
	child := Node{tag, "", make(map[string]*Node), 0, n.Depth + 1}
	n.Children[tag] = &child
	return child
}

//NewNode Simple way to instantiate a Node
func NewNode(tag string) Node {
	return Node{tag, "", make(map[string]*Node), 0, 0}
}

//Add a new text to our json strucutre, by tags
func (n *Node) Add(tags [][]byte, stext string) {
	tag := string(tags[0])

	//If there isn't already a child with this tag, add one
	if _, ok := n.Children[tag]; ok {

	} else {
		n.AppendChild(tag)
	}

	var child *Node
	child = n.Children[tag]
	if len(tags) > 1 {
		child.Add(tags[1:], stext)
	} else {
		child.Text = stext
	}

	n.Elements++
}

//ToJSON turns our struct to a json string
func (n *Node) ToJSON() string {
	//If this is a leaf node, just return the JSON encoded text; parent node returned our tag
	if len(n.Children) == 0 {
		buf := new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		enc.Encode(n.Text)

		//Encoder automatically puts a newline, which is kind of annoying... so we need to remove it
		str := buf.String()

		return str[:len(str)-1]
	}

	//Else, return a JSON object with the sorted tags of children
	out := "{"

	var keys []string
	for k := range n.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		child := n.Children[k]
		json, _ := json.Marshal(child.Tag)
		out += "\n" + strings.Repeat("  ", child.Depth) + string(json) + ":" + child.ToJSON()
		if i < len(keys)-1 {
			out += ","
		}
	}
	out += "\n" + strings.Repeat("  ", n.Depth) + "}"

	return out
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
