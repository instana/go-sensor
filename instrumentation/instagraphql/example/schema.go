// (c) Copyright IBM Corp. 2023

package main

import (
	"errors"

	"github.com/graphql-go/graphql"
)

var chPool = make(map[string][]chan interface{})

var characterType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Character",
	Fields: graphql.Fields{
		"id":         &graphql.Field{Type: graphql.Int},
		"name":       &graphql.Field{Type: graphql.String},
		"profession": &graphql.Field{Type: graphql.String},
		"crewMember": &graphql.Field{Type: graphql.Boolean},
	},
})

var shipType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Ship",
	Fields: graphql.Fields{
		"id":     &graphql.Field{Type: graphql.Int},
		"name":   &graphql.Field{Type: graphql.String},
		"origin": &graphql.Field{Type: graphql.String},
	},
})

func queries(dt *data) graphql.Fields {
	fields := graphql.Fields{
		"characters": &graphql.Field{
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Type: &graphql.List{OfType: characterType},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if id, ok := p.Args["id"].(int); ok {
					return []*character{dt.findChar(id)}, nil
				}

				return dt.Chars, nil
			},
		},
		"ships": &graphql.Field{
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Type: &graphql.List{
				OfType: shipType,
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if id, ok := p.Args["id"].(int); ok {
					return []*ship{dt.findShip(id)}, nil
				}

				return dt.Ships, nil
			},
		},
	}

	return fields
}

func mutations(dt *data) graphql.Fields {
	fields := graphql.Fields{
		"insertCharacter": &graphql.Field{
			Name: "InsertCharacter",
			Type: characterType,
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"profession": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"crewMember": &graphql.ArgumentConfig{
					Type: graphql.Boolean,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var name, profession string
				var ok, crewMember bool

				if name, ok = p.Args["name"].(string); !ok {
					return nil, errors.New("name not found")
				}

				if profession, ok = p.Args["profession"].(string); !ok {
					return nil, errors.New("profession not found")
				}

				if crewMember, ok = p.Args["crewMember"].(bool); !ok {
					return nil, errors.New("crewMember not found")
				}

				c := character{
					Name:       name,
					Profession: profession,
					CrewMember: crewMember,
				}

				dt.addChar(c)

				cName := characterType.Name()

				for idx, ch := range chPool[cName] {
					select {
					case ch <- c:
					default:
						chPool[cName] = append(chPool[cName][:idx], chPool[cName][idx+1:]...)
					}
				}

				return c, nil
			},
		},

		"insertShip": &graphql.Field{
			Name: "InsertShip",
			Type: shipType,
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"origin": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var name, origin string
				var ok bool

				if name, ok = p.Args["name"].(string); !ok {
					return nil, errors.New("name not found")
				}

				if origin, ok = p.Args["origin"].(string); !ok {
					return nil, errors.New("origin not found")
				}

				s := ship{
					Name:   name,
					Origin: origin,
				}

				dt.addShip(s)

				sName := shipType.Name()

				for idx, ch := range chPool[sName] {
					select {
					case ch <- s:
					default:
						chPool[sName] = append(chPool[sName][:idx], chPool[sName][idx+1:]...)
					}
				}

				return s, nil
			},
		},
	}

	return fields
}

func subscriptions(dt *data) graphql.Fields {
	fields := graphql.Fields{
		"newCharacterSubscription": &graphql.Field{
			Name: "character",
			Type: characterType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source, nil
			},
			Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
				ch := make(chan interface{})
				cName := characterType.Name()

				chPool[cName] = append(chPool[cName], ch)

				return ch, nil
			},
		},
		"newShipSubscription": &graphql.Field{
			Name: "ship",
			Type: shipType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source, nil
			},
			Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
				ch := make(chan interface{})
				sName := shipType.Name()

				chPool[sName] = append(chPool[sName], ch)

				return ch, nil
			},
		},
	}

	return fields
}
