// (c) Copyright IBM Corp. 2023

package main

import (
	"encoding/json"
	"io/ioutil"
)

type data struct {
	Chars []character `json:"characters"`
	Ships []ship      `json:"ships"`
}

func (d data) findChar(id int) *character {
	for _, cr := range d.Chars {
		if cr.Id == id {
			return &cr
		}
	}

	return nil
}

func (d *data) addChar(c character) character {
	c.Id = len(d.Chars) + 1
	d.Chars = append(d.Chars, c)
	return c
}

func (d *data) addShip(s ship) ship {
	s.Id = len(d.Ships) + 1
	d.Ships = append(d.Ships, s)
	return s
}

func (d data) findShip(id int) *ship {
	for _, sh := range d.Ships {
		if sh.Id == id {
			return &sh
		}
	}

	return nil
}

func loadData() (*data, error) {
	jsonData, err := ioutil.ReadFile("data.json")

	if err != nil {
		return nil, err
	}

	var dt data

	if err = json.Unmarshal(jsonData, &dt); err != nil {
		return nil, err
	}

	return &dt, nil
}
