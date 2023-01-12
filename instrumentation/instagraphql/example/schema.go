package main

import "github.com/graphql-go/graphql"

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

func queriesWithoutResolve() graphql.Fields {
	fields := graphql.Fields{
		"characters": &graphql.Field{
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Type: &graphql.List{OfType: characterType},
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
		},
	}

	return fields
}

func mutationsWithoutResolve() graphql.Fields {
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
		},
	}

	return fields
}
