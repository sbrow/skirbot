package skirbot

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx context.Context
var client *mongo.Client

func init() {
	uri := `mongodb+srv://user:nsA0oFBd!D4ybfM#@skirmish.ork5l.mongodb.net/skirmish?retryWrites=true&w=majority`

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
}

type (
	Character struct {
		Damage int
		Life   interface{} // Can be number or X
		Speed  int
	}

	Hero struct {
		// Character
		Resolve int
	}

	Halo struct {
		Hero Hero
		Character
		Short string
	}

	Deck struct {
		Leader    string
		Rarity    int
		Cost      interface{} // Cost can be number or X
		Channeled bool
	}

	Card struct {
		Name string

		Continuous bool
		Deck       *Deck
		Hero       *Hero
		Character  *Character
		Halo       *Halo

		Action bool
		Short  string
		Long   string
		Flavor string
	}
)

func (c Character) String() string {
	str := fmt.Sprintf("%v/%v/%v", c.Speed, c.Damage, c.Life)

	pattern := regexp.MustCompile("^1\\/")
	str = pattern.ReplaceAllString(str, "")
	return str
}

func (c Hero) String() string {
	return fmt.Sprintf("{%v}", c.Resolve)
}

func (c Halo) String() string {
	return fmt.Sprintf("//\n%s %s \"%s\"", c.Hero, c.Character, strings.ReplaceAll(c.Short, "\n", " "))
}

func (c Deck) String() string {
	return fmt.Sprintf("%dx[%s] (%v)", c.Rarity, c.Leader, c.Cost)
}

func (c Card) String() string {
	str := fmt.Sprintf("%s %s %s %s %s \"%s\" %s %s", c.Name, c.Deck, c.Hero, c.Character, c.Type(), strings.ReplaceAll(c.Short, "\n", " "), c.Flavor, c.Halo)

	str = strings.ReplaceAll(str, "<nil>", "")
	str = strings.ReplaceAll(str, "\"\"", "")
	str = strings.ReplaceAll(str, "//\n", "//\n"+c.Name+" (Halo) ")
	return removeExtraWhitespace(str)
}

func (c Card) Type() string {
	t := []string{}
	if c.Continuous {
		t = append(t, "Continuous")
	}
	if c.Deck != nil {
		if c.Deck.Channeled {
			t = append(t, "Channeled")
		}
		if c.Character != nil {
			// Some sort of character
			if c.Hero != nil {
				t = append(t, "Hero")
			} else {
				t = append(t, "Follower")
			}
		} else {
			// Actions and events
			if c.Action {
				t = append(t, "Action")
			} else {
				t = append(t, "Event")
			}
		}
	} else {
		t = append(t, "Leader", "Hero")
	}
	return strings.Join(t, " ")
}

func removeExtraWhitespace(s string) string {
	reg := regexp.MustCompile(` +`)
	return reg.ReplaceAllString(s, " ")
}

func GetCardByName(name string) (card *Card, err error) {
	err = client.Database("skirmish").Collection("cards").FindOne(ctx, bson.D{{"name", name}}).Decode(&card)

	return card, err
}
