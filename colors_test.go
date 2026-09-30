package acell

import "testing"

func TestFg256(t *testing.T) {
	t.Run("test with 200", func(t *testing.T) {
		s := Fg256(200)
		expected := "38;5;200"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("test with 2", func(t *testing.T) {
		s := Fg256(2)
		expected := "32"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("test with 8", func(t *testing.T) {
		s := Fg256(8)
		expected := "90"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("test with 11", func(t *testing.T) {
		s := Fg256(11)
		expected := "93"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("negative", func(t *testing.T) {
		s := Fg256(-1)
		expected := "38;5;0"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("too big", func(t *testing.T) {
		s := Fg256(300)
		expected := "38;5;255"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})
}

func TestBg256(t *testing.T) {
	t.Run("test with 200", func(t *testing.T) {
		s := Bg256(200)
		expected := "48;5;200"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("test with 2", func(t *testing.T) {
		s := Bg256(2)
		expected := "42"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("test with 8", func(t *testing.T) {
		s := Bg256(8)
		expected := "100"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("test with 11", func(t *testing.T) {
		s := Bg256(11)
		expected := "103"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("negative", func(t *testing.T) {
		s := Bg256(-1)
		expected := "48;5;0"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})

	t.Run("too big", func(t *testing.T) {
		s := Bg256(300)
		expected := "48;5;255"
		if s != expected {
			t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
		}
	})
}

func TestBgRGB(t *testing.T) {
	s := BgRGB(200, 100, 50)
	expected := "48;2;200;100;50"
	if s != expected {
		t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
	}
}

func TestFgRGB(t *testing.T) {
	s := FgRGB(200, 100, 50)
	expected := "38;2;200;100;50"
	if s != expected {
		t.Fatalf("invalid result: expected %s, but got: %s", expected, s)
	}
}
