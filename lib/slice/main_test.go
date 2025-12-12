package slice

import (
	"reflect"
	"testing"
)

func TestCount(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		element  int
		expected int
	}{
		{
			name:     "empty slice",
			slice:    []int{},
			element:  1,
			expected: 0,
		},
		{
			name:     "no occurrences",
			slice:    []int{1, 2, 3, 4, 5},
			element:  99,
			expected: 0,
		},
		{
			name:     "single occurrence",
			slice:    []int{1, 2, 3, 4, 5},
			element:  3,
			expected: 1,
		},
		{
			name:     "multiple occurrences",
			slice:    []int{1, 2, 3, 2, 5, 2},
			element:  2,
			expected: 3,
		},
		{
			name:     "all elements are target",
			slice:    []int{7, 7, 7, 7},
			element:  7,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Count(tt.slice, tt.element)
			if result != tt.expected {
				t.Errorf("Count() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCountStrings(t *testing.T) {
	slice := []string{"a", "b", "c", "b", "d"}
	result := Count(slice, "b")
	if result != 2 {
		t.Errorf("Count() = %d, want 2", result)
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		delim    int
		expected [][]int
	}{
		{
			name:     "empty slice",
			slice:    []int{},
			delim:    1,
			expected: [][]int{{}},
		},
		{
			name:     "no delimiter",
			slice:    []int{1, 2, 3},
			delim:    99,
			expected: [][]int{{1, 2, 3}},
		},
		{
			name:     "single delimiter in middle",
			slice:    []int{1, 2, 0, 3, 4},
			delim:    0,
			expected: [][]int{{1, 2}, {3, 4}},
		},
		{
			name:     "multiple delimiters",
			slice:    []int{1, 0, 2, 0, 3},
			delim:    0,
			expected: [][]int{{1}, {2}, {3}},
		},
		{
			name:     "delimiter at start",
			slice:    []int{0, 1, 2, 3},
			delim:    0,
			expected: [][]int{{}, {1, 2, 3}},
		},
		{
			name:     "delimiter at end",
			slice:    []int{1, 2, 3, 0},
			delim:    0,
			expected: [][]int{{1, 2, 3}, {}},
		},
		{
			name:     "consecutive delimiters",
			slice:    []int{1, 0, 0, 2},
			delim:    0,
			expected: [][]int{{1}, {}, {2}},
		},
		{
			name:     "all delimiters",
			slice:    []int{0, 0, 0},
			delim:    0,
			expected: [][]int{{}, {}, {}, {}},
		},
		{
			name:     "single element (not delimiter)",
			slice:    []int{1},
			delim:    0,
			expected: [][]int{{1}},
		},
		{
			name:     "single element (is delimiter)",
			slice:    []int{0},
			delim:    0,
			expected: [][]int{{}, {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Split(tt.slice, tt.delim)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Split() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSplitStrings(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		delim    string
		expected [][]string
	}{
		{
			name:     "split sentence by space",
			slice:    []string{"hello", " ", "world", " ", "test"},
			delim:    " ",
			expected: [][]string{{"hello"}, {"world"}, {"test"}},
		},
		{
			name:     "split with newline",
			slice:    []string{"line1", "\n", "line2", "\n", "line3"},
			delim:    "\n",
			expected: [][]string{{"line1"}, {"line2"}, {"line3"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Split(tt.slice, tt.delim)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Split() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSplitBytes tests with byte slices (common use case)
func TestSplitBytes(t *testing.T) {
	input := []byte("hello\nworld\ntest")
	// Convert to slice of bytes for splitting
	result := Split(input, byte('\n'))

	expected := [][]byte{
		[]byte("hello"),
		[]byte("world"),
		[]byte("test"),
	}

	if len(result) != len(expected) {
		t.Fatalf("Split() returned %d parts, want %d", len(result), len(expected))
	}

	for i := range result {
		if !reflect.DeepEqual(result[i], expected[i]) {
			t.Errorf("Split() part %d = %v, want %v", i, result[i], expected[i])
		}
	}
}
