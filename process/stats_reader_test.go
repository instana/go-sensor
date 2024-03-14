// (c) Copyright IBM Corp. 2024

//go:build !linux
// +build !linux

package process

import (
	"reflect"
	"testing"
)

func TestStats(t *testing.T) {
	tests := []struct {
		name string
		want statsReader
	}{
		{
			name: "success",
			want: statsReader{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Stats(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Stats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_statsReader_Memory(t *testing.T) {
	tests := []struct {
		name    string
		want    MemStats
		wantErr bool
	}{
		{
			name:    "success",
			want:    MemStats{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := statsReader{}
			got, err := s.Memory()
			if (err != nil) != tt.wantErr {
				t.Errorf("statsReader.Memory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("statsReader.Memory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_statsReader_CPU(t *testing.T) {
	tests := []struct {
		name    string
		want    CPUStats
		want1   int
		wantErr bool
	}{
		{
			name:    "success",
			want:    CPUStats{},
			want1:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := statsReader{}
			got, got1, err := s.CPU()
			if (err != nil) != tt.wantErr {
				t.Errorf("statsReader.CPU() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("statsReader.CPU() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("statsReader.CPU() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_statsReader_Limits(t *testing.T) {
	tests := []struct {
		name    string
		want    ResourceLimits
		wantErr bool
	}{
		{
			name:    "success",
			want:    ResourceLimits{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := statsReader{}
			got, err := s.Limits()
			if (err != nil) != tt.wantErr {
				t.Errorf("statsReader.Limits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("statsReader.Limits() = %v, want %v", got, tt.want)
			}
		})
	}
}
