// (c) Copyright IBM Corp. 2023

package main

type character struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Profession string `json:"profession"`
	CrewMember bool   `json:"crewMember"`
}

type ship struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Origin string `json:"origin"`
}
