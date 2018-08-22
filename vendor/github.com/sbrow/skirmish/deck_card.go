package skirmish

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// DeckCard is a card that goes in a leader's deck.
type DeckCard struct {
	card
	cost   string
	copies int
}

// NewDeckCard returns a pointer to a new, empty DeckCard.
func NewDeckCard() *DeckCard {
	return &DeckCard{}
}

// Copies returns the number of copies that appear in this card's deck.
func (d *DeckCard) Copies() int {
	return d.copies
}

// SetCard makes c the DeckCard's base card.
func (d *DeckCard) SetCard(c Card) {
	d.card = c.Card()
}

// Cost returns the DeckCard's cost.
// Cost always returns a nil error
func (d *DeckCard) Cost() (string, error) {
	return fmt.Sprint(d.cost), nil
}

// SetCost sets the DeckCard's cost to the given value.
func (d *DeckCard) SetCost(c string) {
	d.cost = c
}

// String returns the string representation of the DeckCard.
func (d *DeckCard) String() string {
	cost, _ := d.Cost() // d.Cost can't return an error.
	s := fmt.Sprintf("%s %dx[%s]", d.card.String(), d.Copies(), d.Leader())
	old := d.Name()
	new := fmt.Sprintf("%s (%s)", d.Name(), cost)
	if d.Resolve() != "" {
		old += " "
	}
	return strings.Replace(s, old, new, 1)
}

// Rarity returns the string representation of d.Copies().
func (d *DeckCard) Rarity() string {
	rarities := map[int]string{
		1: "rare",
		2: "uncommon",
		3: "common",
	}
	return rarities[d.copies]
}

// SetRarity sets the DeckCard's copies to *r. If r is nil,
// d remains unchanged.
func (d *DeckCard) SetRarity(r *int) {
	if r != nil {
		d.copies = *r
	}
}

// Labels returns the column labels to use when marshaling d into to csv format.
func (d *DeckCard) Labels() []string {
	labels := append(d.card.Labels(), "cost", "type", "border_normal", "action",
		"event", "continuous", "item", "show_resolve", "show_speed",
		"show_tough", "show_life", "common", "uncommon", "rare", "rare_border")
	return labels
}

// NormalBorder returns whether or not to show the normal border
// when applying this card to the Photoshop template.
//
// NormalBorder will return false for any card with one of the following attributes:
// 		- "Rare" rarity.
// 		- "Action" card type.
// 		- "Item" card type.
// 		- "Continuous" super type.
func (d *DeckCard) NormalBorder() bool {
	switch {
	case d.copies == 1:
		fallthrough
	case d.Type() == "Action":
		fallthrough
	case d.Type() == "Deck Hero":
		fallthrough
	case d.Type() == "Item":
		fallthrough
	case strings.Contains(strings.Join(d.STypes(), ","), "Continuous"):
		return false
	}
	return true
}

// CSV returns the card in CSV format. If labels is true,
// the first row of the output will be the contents of d.Labels().
func (d *DeckCard) CSV(labels bool) [][]string {
	out := d.card.CSV(true)
	out[0] = d.Labels()
	l := d.Labels()[len(d.card.Labels()):]
	typ := func(label string) string {
		return fmt.Sprint(strings.Contains(d.Type(),
			strings.Title(label)))
	}
	rarity := func(label string) string {
		return fmt.Sprint(d.Rarity() == label)
	}
	labelMap := map[string]string{
		"cost": func() string {
			cost, err := d.Cost()
			if err != nil {
				log.Println(err)
				return ""
			}
			return cost
		}(),
		"type": strings.TrimSpace(fmt.Sprintf("%s %s",
			strings.Join(d.superTypes, " "), d.Type())),
		"action": typ("action"),
		"event":  typ("event"),
		"item":   typ("item"),
		"continuous": fmt.Sprint(strings.Contains(
			strings.Join(d.STypes(), ","), "Continuous")),
		"show_resolve":  fmt.Sprint(d.Resolve() != "0" && d.Resolve() != ""),
		"show_speed":    fmt.Sprint(d.Speed() != 0),
		"show_tough":    fmt.Sprint(strings.Contains(d.Type(), "Follower")),
		"show_life":     fmt.Sprint(strings.Contains(d.Type(), "Hero")),
		"border_normal": fmt.Sprint(d.NormalBorder()),
		"rare_border": fmt.Sprint(d.Rarity() == "rare" &&
			!strings.Contains(strings.Join(d.STypes(), ","), "Continuous")),
		"common":   rarity("common"),
		"uncommon": rarity("uncommon"),
		"rare":     rarity("rare"),
	}
	for _, label := range l {
		out[1] = append(out[1], labelMap[label])
	}
	images, err := d.Images()
	if err != nil {
		log.Println(err)
	}
	tmp := make([][]string, len(images)+1, len(out[0]))
	imgIdx, typeIdx := -1, -1
	for j, col := range out[0] {
		switch col {
		case "card_image":
			imgIdx = j
		case "type":
			typeIdx = j
		}
		if imgIdx != -1 && typeIdx != -1 {
			break
		}
	}
	if imgIdx == -1 {
		log.Panic("card_image not found!")
	}
	if typeIdx == -1 {
		log.Panic("type not found!")
	}
	tmp[0] = out[0]
	tmp[1] = out[1]
	out = tmp
	out[1][0] = fmt.Sprintf("%s_%d", d.name, 1)
	out[1][imgIdx] = images[0]
	if len(images) > 1 {
		out[1][typeIdx] = fmt.Sprintf("%s- %s", out[1][typeIdx],
			strings.TrimSuffix(filepath.Base(images[0]), ".png"))
	}
	for j := 2; j <= len(images); j++ {
		out[j] = make([]string, len(out[j-1]))
		copy(out[j], out[j-1])
		out[j][0] = fmt.Sprintf("%s_%d", d.name, j)
		out[j][imgIdx] = images[j-1]
		out[j][typeIdx] = fmt.Sprintf("%s- %s",
			strings.Split(out[j-1][typeIdx], "-")[0],
			strings.TrimSuffix(filepath.Base(images[j-1]), ".png"))
	}
	if labels {
		return out
	}
	return out[1:]
}

// Images returns any and all paths to Images with this card name.
// Multiple Images may exist to account for cameos, etc.
func (d *DeckCard) Images() (paths []string, err error) {
	// Path to a subfolder, assuming the card has multiple images.
	path := filepath.Join(ImageDir, d.leader, d.Name())
	// If the card does not have a subfolder, check in the main folder for
	// an image file.
	if _, err = os.Stat(path); os.IsNotExist(err) {
		// If found, return it, if not, return an error.
		if _, err = os.Stat(path + ".png"); os.IsNotExist(err) {
			return []string{filepath.Join(ImageDir, DefaultImage)}, fmt.Errorf(`No image found for card '%s'`, d.name)
		}
		return []string{path + ".png"}, nil
	}
	dir, err := ioutil.ReadDir(path + "\\")
	if err != nil {
		return nil, err
	}
	paths = make([]string, len(dir))
	for i, file := range dir {
		paths[i] = filepath.Join(path, file.Name())
	}
	return paths, err
}
