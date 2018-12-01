package resolver

import "github.com/768bit/packr/v2/file"

type Packable interface {
	Pack(name string, f file.File) error
}
