package skirmish

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// cardUEJSON
type cardUEJSON struct {
	Name       string
	Type       string `json:"Class"`
	visual     `json:"Visual"`
	Stats      statsUE
	Abilities  []string
	SystemData systemData
}

// statsUE holds Card data for UE Card objects.
type statsUE struct {
	Life        int `json:"Health_Toughness"`
	Resolve     int `json:"ResolveGain"`
	Cost        int `json:"ResolveCost"`
	CurrResolve int `json:"Resolve"`
	Damage      int `json:"Strength"`
	Speed       int
	Short       string `json:"Action"`
	Long        string `json:"Description"`
	Flavor      string `json:"Quote"`
}

// SystemData holds data relevant to the UE rules engine.
type systemData struct {
	PlayConditions        []string
	InteractionConditions []string
	TurnConditions        []string
}

// visual holds the visual components of a UE Card object.
type visual struct {
	FrontTexture  string
	BackTexture   string
	FrameTexture  string
	DeckIcon      string
	CardTypeIcon  string
	LevelIcon     string
	StatsIcon     string
	FrontMaterial string
	BackMaterial  string
	Resolve       string
}

func newVisual(name, leader string, copies int) *visual {
	img := fmt.Sprintf("0%dx_%s", copies, name)
	return &visual{
		FrontTexture:  fmt.Sprintf("Texture2D'/Game/Textures/Card_Decks/%[1]s/%[2]s.%[2]s'", leader, img),
		BackTexture:   "Texture2D'/Game/Textures/Card_Decks/Common/CardBack.CardBack'",
		FrameTexture:  "None",
		DeckIcon:      "None",
		CardTypeIcon:  "None",
		LevelIcon:     "None",
		StatsIcon:     "None",
		FrontMaterial: "None",
		BackMaterial:  "None",
		Resolve:       "None",
	}
}

func (c card) UEJSON(ident bool) ([]byte, error) {
	obj := cardUEJSON{}
	pre := log.Prefix()
	defer log.SetPrefix(pre)
	obj.Name = c.Name()
	obj.Type = "CTE_" + c.cardType
	resolve, err := strconv.Atoi(c.resolve)
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[%s] ", c.Name()))
		log.Println(err)
	}
	obj.Stats = statsUE{Life: c.Life(), Damage: c.Damage(), Speed: c.Speed(),
		Resolve: resolve, Short: c.Short(), Long: c.Long(), Flavor: c.Flavor()}
	obj.Abilities = make([]string, 0)
	obj.visual = *newVisual(c.Name(), "Common", 1)
	obj.SystemData = systemData{make([]string, 0), make([]string, 0), make([]string, 0)}
	if ident {
		return json.MarshalIndent(obj, "", "\t")
	}
	return json.Marshal(obj)
}

type deckCardUEJSON struct {
	cardUEJSON
	CardName   string
	Leader     string
	Supertypes string `json:"Subclass"`
	Copies     int    `json:"CardCountInDeck"`
}

// UEJSON returns the card formatted into JSON files that
// are readable by Unreal Engine.
func (d DeckCard) UEJSON(ident bool) ([]byte, error) {
	pre := log.Prefix()
	defer log.SetPrefix(pre)
	byt, err := d.card.UEJSON(ident)
	if err != nil {
		log.Panic(err)
	}
	obj := deckCardUEJSON{}
	err = json.Unmarshal(byt, &obj.cardUEJSON)
	if err != nil {
		log.Panic(err)
	}
	cost, err := strconv.Atoi(d.cost)
	if err != nil {
		log.SetPrefix(fmt.Sprintf("[%s] ", d.Name()))
		log.Println(err)
	}

	obj.CardName = d.Name()
	obj.Supertypes = "CTE_" + strings.Join(d.superTypes, "_")
	obj.Name = strings.Replace(d.name, " ", "", -1)
	obj.Leader = d.leader
	obj.Copies = d.copies
	obj.visual = *newVisual(d.name, d.leader, d.copies)
	obj.Stats.Cost = cost

	if ident {
		return json.MarshalIndent(obj, "", "\t")
	}
	return json.Marshal(obj)
}

type nonDeckCardUEJSON struct {
	cardUEJSON
	Faction     string
	DeckCards   string
	Power       bool `json:"bHaveActivePower"`
	ActiveStats statsUE
}

// UEJSON returns the card formatted into JSON files that
// are readable by Unreal Engine.
func (n NonDeckCard) UEJSON(ident bool) ([]byte, error) {
	byt, err := n.card.UEJSON(ident)
	if err != nil {
		log.Panic(err)
	}
	obj := nonDeckCardUEJSON{}
	err = json.Unmarshal(byt, &obj.cardUEJSON)
	if err != nil {
		log.Panic(err)
	}
	obj.Faction = "FE_" + n.faction
	if n.resolveB != nil {
		resolve, err := strconv.Atoi(*n.resolveB)
		if err != nil {
			return []byte{}, err
		}
		obj.ActiveStats.Speed = n.SpeedB()
		obj.ActiveStats.Damage = n.DamageB()
		life, err := strconv.Atoi(n.LifeB())
		if err != nil {
			return []byte{}, err
		}
		obj.ActiveStats.Life = life
		if n.shortB != nil {
			obj.ActiveStats.Short = *n.shortB
		}
		if n.longB != nil {
			obj.ActiveStats.Long = *n.longB
		}
		if n.flavorB != nil {
			obj.ActiveStats.Flavor = *n.flavorB
		}
		obj.ActiveStats.Resolve = resolve
	}
	obj.visual.BackTexture = strings.Replace(obj.visual.BackTexture, "CardBack", fmt.Sprintf("01x_%s_Halo", n.name), -1)
	obj.DeckCards = fmt.Sprintf("DataTable'/Game/Data/%sDeck.%[1]sDeck'", n.name)
	if ident {
		return json.MarshalIndent(obj, "", "\t")
	}
	return json.Marshal(obj)
}
