package main

type Quote struct {
	Id           string   `json:"_id"`
	Content      string   `json:"content"`
	Author       string   `json:"author"`
	Tag          []string `json:"tags"`
	AuthorSlug   string   `json:"authorSlug"`
	Length       int      `json:"length"`
	DateAdded    string   `json:"dateAdded"`
	DateModified string   `json:"dateModified"`
}
