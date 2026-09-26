package modifier

import "testing"

func TestResizerScaleAndCover(t *testing.T) {
	r, err := NewResizer(100, 50)
	if err != nil {
		t.Fatal(err)
	}
	src := Size{Width: 200, Height: 200}
	scaled := r.Scale(src)
	if scaled.Width != 50 || scaled.Height != 50 {
		t.Fatalf("scale got %+v", scaled)
	}
	covered, err := r.Cover(src)
	if err != nil {
		t.Fatal(err)
	}
	if covered.Width != 100 || covered.Height != 100 {
		t.Fatalf("cover got %+v", covered)
	}
	contained, err := r.Contain(src)
	if err != nil {
		t.Fatal(err)
	}
	if contained.Width != 50 || contained.Height != 50 {
		t.Fatalf("contain got %+v", contained)
	}
}

func TestResizeDownDoesNotUpsize(t *testing.T) {
	r, err := NewResizer(400, 400)
	if err != nil {
		t.Fatal(err)
	}
	src := Size{Width: 100, Height: 80}
	got := r.ResizeDown(src)
	if got.Width != 100 || got.Height != 80 {
		t.Fatalf("got %+v", got)
	}
}

func TestPivotPositions(t *testing.T) {
	s := Size{Width: 100, Height: 40}.MovePivot("center", 0, 0)
	if s.Pivot.X != 50 || s.Pivot.Y != 20 {
		t.Fatalf("center %+v", s.Pivot)
	}
	s = Size{Width: 100, Height: 40}.MovePivot("bottom-right", 2, 3)
	if s.Pivot.X != 98 || s.Pivot.Y != 37 {
		t.Fatalf("br %+v", s.Pivot)
	}
	s = Size{Width: 100, Height: 40}.MovePivot("top-left", 1, 2)
	if s.Pivot.X != 1 || s.Pivot.Y != 2 {
		t.Fatalf("tl %+v", s.Pivot)
	}
}

func TestInvalidResizer(t *testing.T) {
	if _, err := NewResizer(-1, 10); err == nil {
		t.Fatal("expected error")
	}
	if _, err := NewResizer(0, 0); err == nil {
		t.Fatal("expected zero-zero error")
	}
	if _, err := NewResizer(0, 10); err == nil {
		t.Fatal("expected zero width error")
	}
	if _, err := SizeFromArgs(0, nil); err == nil {
		t.Fatal("expected zero width")
	}
	if _, err := SizeFromArgs(10, []int{0}); err == nil {
		t.Fatal("expected zero height")
	}
	r, err := SizeFromArgs(400, nil)
	if err != nil || !r.HasW || r.HasH {
		t.Fatalf("auto height resizer %+v %v", r, err)
	}
}
