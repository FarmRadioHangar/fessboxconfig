package scanner

import (
	"io"
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
	src := `
	[section]
	foo=bar
	number=1234
	phone_number=+1234

	; this is a comment

	[section2]
	foo-dash=bar
	`
	s := NewScanner(strings.NewReader(src))
	var tok *Token
	var err error
	for err == nil {
		tok, err = s.Scan()
		if err != nil {
			if err.Error() != io.EOF.Error() {
				t.Fatal(err)
			}
		}
		if tok != nil {
			//fmt.Println(tok.Line, ":", tok.Column, " ", tok.Text)

			switch tok.Line {
			case 1:
				if tok.Column > 0 {
					v := "[section]"
					expect := string(v[tok.Column-1])
					if tok.Text != expect {
						t.Errorf("expected %s fot %s", expect, tok.Text)
					}
				}
			case 2:
				if tok.Column > 0 {
					v := "foo=bar"
					expect := string(v[tok.Column-1])
					if tok.Text != expect {
						t.Errorf("expected %s fot %s", expect, tok.Text)
					}
				}
			case 3:
				if tok.Column > 0 {
					v := "number=1234"
					expect := string(v[tok.Column-1])
					if tok.Text != expect {
						t.Errorf("expected %s fot %s", expect, tok.Text)
					}
				}
			case 4:
				if tok.Column > 0 {
					v := "phone_number=+1234"
					expect := string(v[tok.Column-1])
					if tok.Text != expect {
						t.Errorf("expected %s fot %s", expect, tok.Text)
					}
				}
			case 5:
				if tok.Column > 0 {
					v := "[section]"
					expect := string(v[tok.Column-1])
					if tok.Text != expect {
						t.Errorf("expected %s fot %s", expect, tok.Text)
					}
				}
			case 6:
				if s.column == 1 {
					v := "; this is a comment"
					if tok.Text != v {
						t.Errorf("expected comment  %s got %s", v, tok.Text)
					}
				}
			case 7:
				if tok.Column > 0 {
					v := "[section]"
					expect := string(v[tok.Column-1])
					if tok.Text != expect {
						t.Errorf("expected %s fot %s", expect, tok.Text)
					}
				}
			case 8:
				if tok.Column > 0 {
					v := "[section2]"
					expect := string(v[tok.Column-1])
					if tok.Text != expect {
						t.Errorf("expected %s fot %s", expect, tok.Text)
					}
				}
			case 9:
				if tok.Column > 0 {
					v := "foo-dash=bar"
					expect := string(v[tok.Column-1])
					if tok.Text != expect {
						t.Errorf("expected %s fot %s", expect, tok.Text)
					}
				}
			}
		}
	}
}
