package main

import "errors"

// ErrEmpty when the node has no h2 but this is not an error
var ErrEmpty = errors.New("no text found in dc node")

// Category of the server
type Category string

// CategoryUnknown of the server
const CategoryUnknown Category = "Unknown"

// CategoryNew of the server
const CategoryNew Category = "New"

// CategoryStandard of the server
const CategoryStandard Category = "Standard"

// CategoryPreferred of the server
const CategoryPreferred Category = "Preferred"

// CategoryCongested of the server
const CategoryCongested Category = "Congested"

// WorldStatusURL is the target URL for scraping
const WorldStatusURL = "https://na.finalfantasyxiv.com/lodestone/worldstatus/"
