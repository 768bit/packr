package store

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/movinglater/gotools"
	"github.com/gobuffalo/packr/v2/jam/parser"
	"github.com/gobuffalo/packr/v2/plog"
)

var _ Store = &Legacy{}

type Legacy struct {
	*Disk
	boxes map[string][]legacyBox
}

func NewLegacy() *Legacy {
	return &Legacy{
		Disk:  NewDisk("", ""),
		boxes: map[string][]legacyBox{},
	}
}

func (l *Legacy) Pack(box *parser.Box) error {
	files, err := l.Files(box)
	if err != nil {
		return errors.WithStack(err)
	}

	var fcs []legacyFile

	for _, f := range files {
		n := strings.TrimPrefix(f.Name(), box.AbsPath+string(filepath.Separator))
		c, err := l.prepFile(f)
		if err != nil {
			return errors.WithStack(err)
		}
		fcs = append(fcs, legacyFile{Name: n, Contents: c})
	}

	sort.Slice(fcs, func(a, b int) bool {
		return fcs[a].Name < fcs[b].Name
	})

	lbs := l.boxes[box.PackageDir]
	lbs = append(lbs, legacyBox{
		Box:   box,
		Files: fcs,
	})
	l.boxes[box.PackageDir] = lbs
	return nil

	// run := genny.WetRunner(context.Background())
	// if err := run.WithNew(l.Generator(box)); err != nil {
	// 	return errors.WithStack(err)
	// }
	// run.Logger = plog.Logger
	// return run.Run()
}

func (l *Legacy) prepFile(r io.Reader) (string, error) {
	bb := &bytes.Buffer{}
	if _, err := io.Copy(bb, r); err != nil {
		return "", errors.WithStack(err)
	}
	b, err := json.Marshal(bb.Bytes())
	if err != nil {
		return "", errors.WithStack(err)
	}
	return strings.Replace(string(b), "\"", "\\\"", -1), nil
}

func (l *Legacy) Generator() (*genny.Generator, error) {
	g := genny.New()
	for _, b := range l.boxes {
		if len(b) == 0 {
			continue
		}
		bx := b[0].Box
		pkg := bx.Package
		opts := map[string]interface{}{
			"Package": pkg,
			"Boxes":   b,
		}
		f := genny.NewFile(filepath.Join(bx.PackageDir, "a_"+bx.Package+"-packr.go.tmpl"), strings.NewReader(legacyTmpl))
		t := gotools.TemplateTransformer(opts, nil)
		f, err := t.Transform(f)
		if err != nil {
			return g, errors.WithStack(err)
		}
		g.File(f)
	}
	return g, nil
}

func (l *Legacy) Close() error {
	run := genny.WetRunner(context.Background())
	if err := run.WithNew(l.Generator()); err != nil {
		return errors.WithStack(err)
	}
	run.Logger = plog.Logger
	return run.Run()
}

type legacyBox struct {
	Box   *parser.Box
	Files []legacyFile
}

type legacyFile struct {
	Name     string
	Contents string
}

var legacyTmpl = `// Code generated by github.com/gobuffalo/packr. DO NOT EDIT.

package {{.Package}}

import "github.com/768bit/packr"

// You can use the "packr clean" command to clean up this,
// and any other packr generated files.
func init() {
	{{- range $box := .Boxes }}
	{{- range $box.Files }}
	packr.PackJSONBytes("{{$box.Box.Name}}", "{{.Name}}", "{{.Contents}}")
	{{- end }}
	{{- end }}
}
`
