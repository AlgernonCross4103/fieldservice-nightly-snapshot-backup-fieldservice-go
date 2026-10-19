package main

import "testing"

func TestShouldUpload(t *testing.T) {
	tests := []struct {
		name   string
		orders []WorkOrder
		want   bool
	}{{"empty shift", nil, false}, {"completed work", []WorkOrder{{ID: "WO-1"}}, true}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldUpload(tt.orders); got != tt.want {
				t.Fatalf("shouldUpload() = %v, want %v", got, tt.want)
			}
		})
	}
}
