package store

import (
	"github.com/768bit/packr/v2/jam/parser"
)

type Store interface {
	FileNames(*parser.Box) ([]string, error)
	Files(*parser.Box) ([]*parser.File, error)
	Pack(*parser.Box) error
	Clean(*parser.Box) error
}
