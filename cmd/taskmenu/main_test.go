package main

import "testing"

func TestMoveCursorWraps(t *testing.T) {
	st := &menuState{}

	st.moveCursor(-1, 5)
	if st.cursor != 4 {
		t.Fatalf("up from first: cursor = %d, want 4", st.cursor)
	}

	st.moveCursor(1, 5)
	if st.cursor != 0 {
		t.Fatalf("down from last: cursor = %d, want 0", st.cursor)
	}

	st.moveCursor(1, 5)
	if st.cursor != 1 {
		t.Fatalf("down from first: cursor = %d, want 1", st.cursor)
	}

	st.moveCursor(-1, 5)
	if st.cursor != 0 {
		t.Fatalf("up from second: cursor = %d, want 0", st.cursor)
	}
}

func TestMoveCursorEmpty(t *testing.T) {
	st := &menuState{}
	st.moveCursor(1, 0)
	st.moveCursor(-1, 0)
	if st.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", st.cursor)
	}
}

func TestScrollWindowFollowsCursor(t *testing.T) {
	const total, rows = 27, 19
	st := &menuState{}

	check := func(cursor, wantOffset int) {
		t.Helper()
		st.cursor = cursor
		st.scrollWindow(total, rows)
		if st.offset != wantOffset {
			t.Fatalf("cursor %d: offset = %d, want %d", cursor, st.offset, wantOffset)
		}
		if st.cursor < st.offset || st.cursor >= st.offset+rows {
			t.Fatalf("cursor %d out of window [%d,%d)", st.cursor, st.offset, st.offset+rows)
		}
	}

	check(0, 0)
	check(18, 0)
	check(19, 1)
	check(26, total-rows)
	check(26, total-rows)
	check(7, 7)
	check(0, 0)
}

func TestScrollWindowShortList(t *testing.T) {
	st := &menuState{}
	st.cursor = 3
	st.scrollWindow(4, 19)
	if st.offset != 0 {
		t.Fatalf("offset = %d, want 0", st.offset)
	}
}

func TestBuildSelectionUsesCursor(t *testing.T) {
	st := &menuState{
		tasks:    []Task{{Name: "build"}, {Name: "test"}},
		cursor:   1,
		selected: map[int]bool{},
	}

	got, ok := st.buildSelection()
	if !ok || len(got) != 1 || got[0] != 1 {
		t.Fatalf("selection = %v, ok = %v, want [1] true", got, ok)
	}
}

func TestSelectionNumbersRenumber(t *testing.T) {
	st := &menuState{selected: map[int]bool{}}
	st.toggleSelected(14)
	st.toggleSelected(1)
	st.toggleSelected(7)

	numbers := st.selectionNumbers()
	if numbers[14] != 1 || numbers[1] != 2 || numbers[7] != 3 {
		t.Fatalf("numbers = %v, want 14:1 1:2 7:3", numbers)
	}

	st.toggleSelected(1)
	numbers = st.selectionNumbers()
	if _, ok := numbers[1]; ok {
		t.Fatalf("deselected task still numbered: %v", numbers)
	}
	if numbers[14] != 1 || numbers[7] != 2 {
		t.Fatalf("numbers after deselect = %v, want 14:1 7:2", numbers)
	}
}

func TestResetKeepsCursorPosition(t *testing.T) {
	st := &menuState{selected: map[int]bool{}}
	st.cursor = 4
	st.offset = 2
	st.toggleSelected(3)
	st.notice = "done"

	st.reset()

	if st.cursor != 4 || st.offset != 2 {
		t.Fatalf("after reset: cursor = %d, offset = %d, want 4, 2", st.cursor, st.offset)
	}
	if len(st.selectedIndices()) != 0 {
		t.Fatalf("selection not cleared: %v", st.selectedIndices())
	}
	if st.notice != "" {
		t.Fatalf("notice not cleared: %q", st.notice)
	}
}

func TestSelectionPreservesOrder(t *testing.T) {
	st := &menuState{selected: map[int]bool{}}

	st.toggleSelected(14)
	st.toggleSelected(1)
	st.toggleSelected(7)

	got := st.selectedIndices()
	want := []int{14, 1, 7}
	if len(got) != len(want) {
		t.Fatalf("selection = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("selection = %v, want %v", got, want)
		}
	}

	st.toggleSelected(1)
	got = st.selectedIndices()
	want = []int{14, 7}
	if len(got) != len(want) || got[0] != 14 || got[1] != 7 {
		t.Fatalf("after deselect: selection = %v, want %v", got, want)
	}

	st.toggleSelected(1)
	got = st.selectedIndices()
	want = []int{14, 7, 1}
	if len(got) != len(want) || got[0] != 14 || got[1] != 7 || got[2] != 1 {
		t.Fatalf("after reselect: selection = %v, want %v", got, want)
	}
}
