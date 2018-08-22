package skirmish

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

// Card is an interface shared by DeckCards and NonDeckCards.
//
// It is important to note that each Card object can represent more than one card.
// 		- DeckCards can hold multiple cards with unique art.
//		- NonDeckCards can hold their frontside Card and rearside Card.
//
// Setters
//
// The Set methods accept pointers because SQL queries return pointer values.
// Passing nil to a set value will generally result in no change to the Card.
//
// TODO(sbrow): Figure out how to handle id. [Issue](https://github.com/sbrow/skirmish/issues/33)
//
// TODO(sbrow): Cameo card flavor text. [Issue](https://github.com/sbrow/skirmish/issues/32)
type Card interface {
	Card() card
	Faction() string
	ID(int) string
	Name() string
	FullType() string
	Resolve() string
	STypes() []string
	Speed() int
	Type() string
	Copies() int

	SetDamage(*int)
	SetFlavor(*string)
	SetLeader(*string)
	SetLife(*int)
	SetLong(*string)
	SetName(*string)
	SetRegexp(*string)
	SetResolve(*string)
	SetShort(*string)
	SetSpeed(*int)
	SetSTypes([]string)
	SetType(*string)

	Damage() int
	Leader() string
	Life() int
	Short() string
	Regexp() string
	Long() string
	Flavor() string
	Labels() []string
	String() string
	Images() ([]string, error)
	Bold() ([][]int, error)
	MarshalXML(*xml.Encoder, xml.StartElement) error
	UEJSON(ident bool) ([]byte, error)
	CSV(bool) [][]string
	XML() ([]byte, error)
}

// Load queries the standard database for a card with the given name, and returns the
// result as a Card object.
func Load(name string) (Card, error) {
	cards, err := LoadMany(fmt.Sprintf("name='%s'", strings.Replace(name, "'", "''", -1)))
	if len(cards) > 0 {
		return cards[0], err
	}
	if err != nil {
		return nil, err
	}
	return nil, errors.New("No card found with name " + name + ", check your spelling.")
}

var props = []string{"\"name\"", "cards.type", "cards.supertypes",
	"cards.short", "cards.long", "flavor", "resolve", "cards.speed", "cards.damage",
	"cards.life", "cards.faction", "cards.cost", "cards.rarity", "cards.leader",
	"cards.resolve_b", "cards.life_b", "cards.speed_b", "cards.damage_b",
	"cards.short_b", "cards.long_b", "cards.flavor_b", "cards.regexp"}

// LoadMany queries the standard database for all cards that match the given condition
// and returns them as a slice of Card objects.
func LoadMany(cond string) ([]Card, error) {
	if db == nil {
		if err := Connect(LocalDB.DBArgs()); err != nil {
			return []Card{}, err
		}
	}
	if cond == "" {
		cond = "1"
	}
	out := make([]Card, 0)
	str := fmt.Sprintf("select %s from cards WHERE %s",
		strings.Join(props, ", "), cond)
	rows, err := Query(str)
	if err != nil {
		return []Card{}, err
	}
	for rows.Next() {
		var typ, supertypes, title, short, long, flavor, resolve, faction, leader,
			resolveB, lifeB, shortB, longB, flavorB, cost, regexp *string
		var speed, damage, life, rarity, speedB, damageB *int
		err := rows.Scan(&title, &typ, &supertypes, &short, &long,
			&flavor, &resolve, &speed, &damage, &life, &faction, &cost, &rarity,
			&leader, &resolveB, &lifeB, &speedB, &damageB, &shortB, &longB, &flavorB,
			&regexp)
		c := NewCard()
		if err != nil {
			return out, err
		}
		c.SetType(typ)
		if supertypes != nil {
			// TODO(sbrow): Figure out how to pass a pointer to card.SetSTypes [Issue](https://github.com/sbrow/skirmish/issues/31)
			c.SetSTypes(strings.Split(*supertypes, ","))
		}
		c.SetName(title)
		c.SetShort(short)
		c.SetLong(long)
		c.SetFlavor(flavor)
		c.SetResolve(resolve)
		c.SetSpeed(speed)
		c.SetDamage(damage)
		c.SetLife(life)
		c.SetRegexp(regexp)
		switch {
		case cost != nil:
			d := &DeckCard{}
			d.SetCard(c)
			d.SetCost(*cost)
			d.SetRarity(rarity)
			d.SetLeader(leader)
			out = append(out, d)
		case supertypes != nil && strings.Contains(*supertypes, "Leader"):
			n := &NonDeckCard{}
			n.SetCard(c)
			n.SetResolveB(resolveB)
			n.SetLifeB(lifeB)
			n.SetSpeedB(speedB)
			n.SetDamageB(damageB)
			n.SetShortB(shortB)
			n.SetLongB(longB)
			n.SetFlavorB(flavorB)
			n.SetFaction(faction)
			out = append(out, n)
		default:
			c.SetLeader(leader)
			out = append(out, c)
		}
	}
	return out, nil
}

// NewCard returns a new, empty Card object.
func NewCard() Card {
	return &card{}
}

// Card is the base struct for DeckCards and NonDeckCards.
type card struct {
	name       string   // The name of the card.
	leader     string   // The name of the card's leader.
	cardType   string   // The card's type.
	superTypes []string // The card's supertype(s).
	resolve    string   // The resolve this card produces when in play.
	stats      stats    // The card's speed, life, and damage, if applicable.
	short      string   // The card's basic rules text.
	long       string   // The card's reminder text.
	flavor     string   // The card's flavor (non-rules) text.
	regexp     string   // A regular expression for what characters should be bold in short.
}

func (c *card) Bold() ([][]int, error) {
	reg, err := regexp.Compile(c.regexp)
	if err != nil {
		fmt.Println(c.regexp)
		return [][]int{}, err
	}
	return reg.FindAllStringIndex(c.short, -1), nil
}

// Copies returns the number of times this card appears in the game.
func (c *card) Copies() int {
	return 1
}

// Delim is the Delimiter to use when Marshalling cards to csv format.
var Delim = ","

// CSV returns the card in CSV format. If labels is true,
// the first row of the output will be the contents of card.Labels().
func (c *card) CSV(labels bool) [][]string {
	str := make([]string, len(c.Labels()))
	space := func(str string) string {
		if len(str) == 0 {
			return " "
		}
		return str
	}
	labelMap := map[string]string{
		"name": c.Name(),
		"resolve": func() string {
			if c.Resolve() == "" {
				return "0"
			}
			return c.Resolve()
		}(),
		"speed":  fmt.Sprint(c.Speed()),
		"damage": fmt.Sprint(c.Damage()),
		"life":   fmt.Sprint(c.Life()),
		"short":  space(c.Short()),
		"long":   space(c.Long()),
		"flavor": space(c.Flavor()),
		"card_image": func() string {
			img, err := c.Images()
			if err != nil {
				log.Println(err)
			}
			return img[0]
		}(),
	}

	for i, label := range c.Labels() {
		str[i] += labelMap[label]
		if strings.Contains(strings.Join(Leaders.names(), ","), label) {
			str[i] += fmt.Sprint(c.Leader() == label)
		}
		str[i] += Delim
	}
	for i := range str {
		str[i] = strings.TrimSuffix(str[i], Delim)
	}
	if labels {
		return [][]string{c.Labels(), str}
	}
	return [][]string{str}
}

func (c *card) Card() card {
	return *c
}

func (c *card) Damage() int {
	return c.stats.damage
}
func (c *card) Life() int {
	return c.stats.life
}

func (c *card) Faction() string {
	return ""
}

func (c *card) Name() string {
	return c.name
}

func (c *card) Resolve() string {
	return fmt.Sprint(c.resolve)
}

func (c *card) Speed() int {
	return c.stats.speed
}

func (c *card) SetDamage(d *int) {
	if d != nil {
		c.stats.damage = *d
	}
}

func (c *card) SetLife(l *int) {
	if l != nil {
		c.stats.life = *l
	}
}

func (c *card) SetName(name *string) {
	if name != nil {
		c.name = *name
	}
}

func (c *card) SetResolve(r *string) {
	if r == nil {
		return
	}
	m, err := regexp.Match(`[+\-][1-9]`, []byte(*r))
	if err != nil {
		log.Panic(err)
	}
	if m {
		c.resolve = *r
	}
}

func (c *card) SetSpeed(s *int) {
	if s != nil {
		c.stats.speed = *s
	}
}

func (c *card) Short() string {
	return c.short
}

func (c *card) SetShort(s *string) {
	if s != nil {
		c.short = *s
	}
}

func (c *card) Long() string {
	return c.long
}

func (c *card) SetLong(s *string) {
	if s != nil {
		c.long = *s
	}
}

func (c *card) Flavor() string {
	return c.flavor
}

func (c *card) SetFlavor(s *string) {
	if s != nil {
		c.flavor = *s
	}
}

func (c *card) Type() string {
	return c.cardType
}

func (c *card) SetType(t *string) {
	if t != nil {
		c.cardType = *t
	}
}

func (c *card) STypes() []string {
	return c.superTypes
}

func (c *card) SetSTypes(t []string) {
	c.superTypes = t
}

func (c *card) Leader() string {
	return c.leader
}

func (c *card) SetLeader(l *string) {
	if l != nil {
		c.leader = *l
	}
}

func (c *card) SetRegexp(reg *string) {
	if reg != nil {
		c.regexp = *reg
	}
}

// Regexp returns the regular expression that matches
// bold words in the card's short text.
func (c *card) Regexp() string {
	return c.regexp
}

// ID returns an id unique to the card.
//
// Cards with only one art will have identical ID and Name.
// Cards with more than one art will have an ID containing their name
// and which version of art they use.
func (c *card) ID(ver int) string {
	return fmt.Sprintf("%s_%d", c.name, ver)
}

// Labels prints the column labels for .csv output.
func (c card) Labels() []string {
	labels := []string{
		"id", "name", "resolve", "speed", "damage", "life",
		"short", "long", "flavor", "card_image",
	}
	if len(Leaders) == 0 {
		log.Fatal("No leaders found when computing labels")
	}
	labels = append(labels, Leaders.names()...)
	return labels
}

// Image returns the path to the card's illustration.
// Images for deck cards must be in a directory named after their leader.
// Images for leader and partner heroes should be in a directory names "Heroes".
//
// path = [$SK_SRC]/[folder]]/[c.Name].png
func (c *card) Images() (paths []string, err error) {
	return []string{filepath.Join(ImageDir, "ImageNotFound.png")}, nil
}

func (c *card) JSON() ([]byte, error) {
	bytes, err := json.Marshal(c)
	return bytes, err
}

func (c *card) FullType() string {
	if len(c.STypes()) == 0 {
		return c.cardType
	}
	return fmt.Sprintf("%s %s", strings.Join(c.superTypes, " "), c.cardType)
}

func (c *card) String() string {
	return noSpaces(fmt.Sprintf("%s {%s} %s %s- \"%s\"", c.Name(), c.Resolve(),
		c.stats.String(), c.FullType(), pruneNewLines(c.short)))
}

func noSpaces(s string) string {
	reg := regexp.MustCompile(`\s\s+`)
	return reg.ReplaceAllString(s, " ")
}
func pruneNewLines(s string) string {
	reg := regexp.MustCompile("[\r\n]+")
	return reg.ReplaceAllString(s, " ")
}

// stats holds a card's character stats.
type stats struct {
	speed  int
	damage int
	life   int
}

func (s stats) String() string {
	reg := regexp.MustCompile(`^[0-1]\/(0\/)*0?`)
	str := fmt.Sprintf("%d/%d/%d", s.speed, s.damage, s.life)
	return reg.ReplaceAllString(str, "")
}
