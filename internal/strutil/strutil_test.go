package strutil

import (
	"reflect"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "contains item in middle",
			slice: []string{"a", "b", "c"},
			item:  "b",
			want:  true,
		},
		{
			name:  "contains item at start",
			slice: []string{"a", "b", "c"},
			item:  "a",
			want:  true,
		},
		{
			name:  "contains item at end",
			slice: []string{"a", "b", "c"},
			item:  "c",
			want:  true,
		},
		{
			name:  "does not contain item",
			slice: []string{"a", "b", "c"},
			item:  "d",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "a",
			want:  false,
		},
		{
			name:  "single item match",
			slice: []string{"a"},
			item:  "a",
			want:  true,
		},
		{
			name:  "single item no match",
			slice: []string{"a"},
			item:  "b",
			want:  false,
		},
		{
			name:  "duplicate items",
			slice: []string{"a", "b", "a", "c"},
			item:  "a",
			want:  true,
		},
		{
			name:  "empty string item",
			slice: []string{"a", "", "b"},
			item:  "",
			want:  true,
		},
		{
			name:  "empty string item not found",
			slice: []string{"a", "b"},
			item:  "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("Contains(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.want)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  []string
	}{
		{
			name:  "remove item from middle",
			slice: []string{"a", "b", "c"},
			item:  "b",
			want:  []string{"a", "c"},
		},
		{
			name:  "remove item from start",
			slice: []string{"a", "b", "c"},
			item:  "a",
			want:  []string{"b", "c"},
		},
		{
			name:  "remove item from end",
			slice: []string{"a", "b", "c"},
			item:  "c",
			want:  []string{"a", "b"},
		},
		{
			name:  "remove non-existent item",
			slice: []string{"a", "b", "c"},
			item:  "d",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "a",
			want:  []string{},
		},
		{
			name:  "single item match",
			slice: []string{"a"},
			item:  "a",
			want:  []string{},
		},
		{
			name:  "single item no match",
			slice: []string{"a"},
			item:  "b",
			want:  []string{"a"},
		},
		{
			name:  "remove all duplicates",
			slice: []string{"a", "b", "a", "c", "a"},
			item:  "a",
			want:  []string{"b", "c"},
		},
		{
			name:  "remove empty string",
			slice: []string{"a", "", "b", ""},
			item:  "",
			want:  []string{"a", "b"},
		},
		{
			name:  "remove from slice with only duplicates",
			slice: []string{"a", "a", "a"},
			item:  "a",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Remove(tt.slice, tt.item)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Remove(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.want)
			}
		})
	}
}
