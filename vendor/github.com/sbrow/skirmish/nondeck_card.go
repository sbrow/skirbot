package skirmish

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// NonDeckCard represents a Leader or Partner card.
type NonDeckCard struct {
	card
	faction  string
	resolveB *string
	statsB   *stats
	shortB   *string
	longB    *string
	flavorB  *string
}

// DamageB returns n's Halo side damage.
//
// If n doesn't have a Halo side, DamageB returns 0.
func (n *NonDeckCard) DamageB() int {
	if n.statsB == nil {
		return 0
	}
	return (*n.statsB).damage
}

// LifeB returns n's Halo side life.
//
// If n doesn't have a Halo side, LifeB returns 0.
func (n *NonDeckCard) LifeB() string {
	if n.statsB == nil {
		return ""
	}
	life := (*n.statsB).life
	if life >= 0 {
		return fmt.Sprintf("+%d", life)
	}
	return fmt.Sprintf("-%d", life)
}

// ID returns this card's unique identifier.
func (n *NonDeckCard) ID(ver int) string {
	id := n.Name()
	if ver != 1 {
		id += " (Halo)"
	}
	return id
}

// Faction returns the faction that n is aligned to.
func (n *NonDeckCard) Faction() string {
	return n.faction
}

// ResolveB returns n's Halo side Resolve.
//
// If n doesn't have a Halo side, ResolveB returns "".
func (n *NonDeckCard) ResolveB() string {
	if n.resolveB == nil {
		return ""
	}
	return fmt.Sprintf("+%s", *n.resolveB)
}

// SpeedB returns n's Halo side speed.
//
// If n doesn't have a Halo side, SpeedB returns 0.
func (n *NonDeckCard) SpeedB() int {
	if n.statsB == nil {
		return 0
	}
	return (*n.statsB).speed
}

// ShortB returns n's Halo side short text.
//
// If n doesn't have a Halo side, ShortB returns 0.
func (n *NonDeckCard) ShortB() string {
	if n.shortB == nil {
		return ""
	}
	return *n.shortB
}

// LongB returns n's Halo side long text.
//
// If n doesn't have a Halo side, LongB returns 0.
func (n *NonDeckCard) LongB() string {
	if n.longB == nil {
		return ""
	}
	return *n.longB
}

// SetDamageB sets n's Halo side damage to d.
//
// If n doesn't have a Halo side, or if d is nil, SetDamageB does nothing.
func (n *NonDeckCard) SetDamageB(d *int) {
	if n.statsB == nil {
		n.statsB = &stats{}
	}
	if d != nil {
		n.statsB.damage = *d
	}
}

// SetFaction sets n's Faction to faction.SetFaction
//
// If faction is nil, n is not changed.
func (n *NonDeckCard) SetFaction(faction *string) {
	if faction != nil {
		n.faction = *faction
	}
}

// SetFlavorB sets n's Halo side flavor to f.
//
// If n doesn't have a Halo side, or if f is nil, SetFlavorB does nothing.
func (n *NonDeckCard) SetFlavorB(f *string) {
	n.flavorB = f
}

// SetLifeB sets n's Halo side life to l.
//
// If n doesn't have a Halo side, or if l is nil, SetLifeB does nothing.
func (n *NonDeckCard) SetLifeB(l *string) {
	if n.statsB == nil {
		n.statsB = &stats{}
	}
	if l != nil {
		life, err := strconv.Atoi(*l)
		if err != nil {
			log.Println(err)
		} else {
			n.statsB.life = life
		}
	}
}

// SetLongB sets n's Halo side long text to l.
//
// If n doesn't have a Halo side, or if l is nil, SetLongB does nothing.
func (n *NonDeckCard) SetLongB(l *string) {
	n.longB = l
}

// SetResolveB sets n's Halo side resolve to r.
//
// If n doesn't have a Halo side, or if r is nil, SetResolveB does nothing.
func (n *NonDeckCard) SetResolveB(r *string) {
	n.resolveB = r
}

// SetShortB sets n's Halo side short text to s.
//
// If n doesn't have a Halo side, or if s is nil, SetShortB does nothing.
func (n *NonDeckCard) SetShortB(s *string) {
	n.shortB = s
}

// SetSpeedB sets n's Halo side speed to s.
//
// If n doesn't have a Halo side, or if s is nil, SetDamageB does nothing.
func (n *NonDeckCard) SetSpeedB(s *int) {
	if n.statsB == nil {
		n.statsB = &stats{}
	}
	if s != nil {
		n.statsB.speed = *s
	}
}

// StatsB returns the string representation of n's Halo side stats.
//
// If n doesn't have a Halo side, an empty string is returned.
func (n *NonDeckCard) StatsB() string {
	reg := regexp.MustCompile(`(\/)([^\/-])*$`)
	return reg.ReplaceAllString(n.statsB.String(), "/+$2")
}

// String returns the string representation of n.
func (n *NonDeckCard) String() string {
	str := n.card.String()
	if n.ResolveB() != "" {
		str += fmt.Sprintf(" //\n%s (Halo) %s %s %s- \"%s\"", n.card.Name(), n.ResolveB(),
			n.StatsB(), n.card.FullType(), pruneNewLines(n.ShortB()))
	}
	return str
}

// Images returns the path's to n's front side and Halo side images (if applicable).
func (n *NonDeckCard) Images() (paths []string, err error) {
	basePath := filepath.Join(ImageDir, "Heroes", n.Name())
	images := []string{basePath + ".png", basePath + " (Halo).png"}
	for i, img := range images {
		if _, err = os.Stat(img); os.IsNotExist(err) {
			images[i] = filepath.Join(ImageDir, DefaultImage)
		}
	}
	if images[1] == filepath.Join(ImageDir, DefaultImage) {
		return []string{images[0]}, nil
	}
	return images, nil
}

// SetCard makes c the DeckCard's base card.
func (n *NonDeckCard) SetCard(c Card) {
	n.card = c.Card()
}

// Labels returns the column labels to use when marshaling n into to csv format.
func (n *NonDeckCard) Labels() []string {
	labels := append(n.card.Labels(), "Halo", "Troika", "Nightmares")
	return labels
}

// CSV returns the card in CSV format. If labels is true,
// the first row of the output will be the contents of n.Labels().
func (n *NonDeckCard) CSV(labels bool) [][]string {
	out := n.card.CSV(true)
	out[0] = n.Labels()
	l := n.Labels()[len(n.card.Labels()):]
	faction := func(label string) string {
		return fmt.Sprint(n.Faction() == label)
	}
	labelMap := map[string]string{
		"Halo":       "false",
		"Troika":     faction("Troika"),
		"Nightmares": faction("Nightmares"),
	}
	for _, label := range l {
		out[1] = append(out[1], labelMap[label])
	}
	images, err := n.Images()
	if err != nil {
		log.Println(err)
	}
	tmp := make([][]string, len(images)+1, len(out[0]))
	tmp[0] = out[0]
	tmp[1] = out[1]
	i := -1
	for j, col := range out[0] {
		if col == "card_image" {
			i = j
			break
		}
	}
	out = tmp
	out[1][0] = n.name
	out[1][1] = fmt.Sprintf("%s- %s", n.Type(), n.name)
	out[1][2] = string(out[1][2][1])
	out[1][i] = images[0]
	// out[1][20] = "true"
	for j := 2; j <= len(images); j++ {
		out[j] = make([]string, len(out[j-1]))
		copy(out[j], out[j-1])
		out[j][0] = fmt.Sprintf("%s (Halo)", n.name)
		out[j][1] = fmt.Sprintf("%s- %s", n.Type(), n.name)
		out[j][2] = n.ResolveB()
		out[j][3] = fmt.Sprint(n.SpeedB())
		out[j][4] = fmt.Sprint(n.DamageB())
		out[j][5] = n.LifeB()
		out[j][6] = fmt.Sprintf(`"%s"`, n.ShortB())
		out[j][7] = fmt.Sprintf(`"%s"`, n.LongB())
		out[j][i] = images[j-1]
		out[j][20] = "true"
	}
	if labels {
		return out
	}
	return out[1:]
}
