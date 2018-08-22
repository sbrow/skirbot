package skirmish

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/sbrow/prob/combin"
)

type cardXML struct {
	XMLName xml.Name `xml:"card"`
	Name    string   `xml:"name"`
	Set     string   `xml:"set"`
	Type    string   `xml:"type"`
	PT      string   `xml:"pt,omitempty"`
	Row     uint8    `xml:"row"`
	Text    string   `xml:"text,omitempty"`
}

func (c *card) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	obj := cardXML{
		Name: c.name,
		Set:  "DKC",
		Type: c.FullType(),
		Row:  2,
		Text: fmt.Sprintf("%s\n(%s)", c.Short(), c.Long()),
	}
	if c.Damage() != 0 && c.Life() != 0 {
		obj.PT = fmt.Sprintf("%d/%d", c.Damage(), c.Life())
	}
	for _, t := range c.STypes() {
		if t == "Continuous" {
			obj.Row = 1
		}
	}
	if obj.Row == 2 && c.Type() == "Action" || c.Type() == "Event" {
		obj.Row = 3
	}
	if c.Speed() > 0 {
		obj.Text = fmt.Sprintf("%d Speed.\n%s", c.Speed(), obj.Text)
	}
	return enc.EncodeElement(obj, start)
}

func (c *card) XML() ([]byte, error) {
	return []byte{}, nil // return c.MarshalXML()
}

func (c *DeckCard) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	obj := cardXML{
		Name: c.name,
		Set:  "DKC",
		Type: c.FullType(),
		Row:  2,
		Text: fmt.Sprintf("%s\n(%s)", c.Short(), c.Long()),
	}
	if c.Damage() != 0 && c.Life() != 0 {
		obj.PT = fmt.Sprintf("%d/%d", c.Damage(), c.Life())
	}
	for _, t := range c.STypes() {
		if t == "Continuous" {
			obj.Row = 1
		}
	}
	if obj.Row == 2 && c.Type() == "Action" || c.Type() == "Event" {
		obj.Row = 3
	}
	if c.Speed() > 0 {
		obj.Text = fmt.Sprintf("%d Speed.\n%s", c.Speed(), obj.Text)
	}
	return enc.EncodeElement(obj, start)
}
func decks() {
	str :=
		`<?xml version="1.0" encoding="UTF-8"?>
<cockatrice_deck version="1">
	<deckname></deckname>
	<comments></comments>
	<zone name="main">
`
	light := combin.NewSet("Bast", "Igrath", "Lilith", "Vi", "Scinter").Combine()
	queries := []string{}
	for _, combo := range light {
		queries = append(queries,
			fmt.Sprintf("%s|%s",
				combo[0].(string), combo[1].(string)))
	}
	rows, err := Query("Select cards.rarity, name from cards where cards.leader ~ $1 ORDER BY name ASC", queries[0])
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var name string
		var copies int
		rows.Scan(&copies, &name)
		str += fmt.Sprintf("\t\t<card number=\"%d\" name=\"%s\"/>\n", copies, name)
	}
	str += "\t</zone>\n\t<zone name=\"side\">\n"
	rows, err = Query("SELECT name from cards where name ~ $1", queries[0])
	if err != nil {
		log.Println("error:", err)
	}
	for rows.Next() {
		var name string
		rows.Scan(&name)
		str += fmt.Sprintf("\t\t<card number=\"1\" name=\"%s\"/>\n", name)
		str += fmt.Sprintf("\t\t<card number=\"1\" name=\"%s (Halo)\"/>\n", name)
	}

	str += "\t</zone>\n</cockatrice_deck>"
	fmt.Println(str)
}
