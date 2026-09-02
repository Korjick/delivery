package kernel

import (
	"reflect"
	"testing"
)

func TestLocation_Distance(t *testing.T) {
	type fields struct {
		X int
		Y int
	}
	type args struct {
		o *Location
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		err    error
	}{
		{
			name: "correct_distance_greater_x_y",
			want: 4,
			fields: fields{
				X: 2,
				Y: 3,
			},
			args: args{
				o: &Location{x: 4, y: 5},
			},
		},
		{
			name: "correct_distance_greater_x",
			want: 2,
			fields: fields{
				X: 2,
				Y: 3,
			},
			args: args{
				o: &Location{x: 4, y: 3},
			},
		},
		{
			name: "correct_distance_greater_y",
			want: 2,
			fields: fields{
				X: 2,
				Y: 3,
			},
			args: args{
				o: &Location{x: 2, y: 5},
			},
		},
		{
			name: "correct_distance_lower_x_y",
			want: 4,
			fields: fields{
				X: 4,
				Y: 5,
			},
			args: args{
				o: &Location{x: 2, y: 3},
			},
		},
		{
			name: "correct_distance_lower_x",
			want: 2,
			fields: fields{
				X: 4,
				Y: 3,
			},
			args: args{
				o: &Location{x: 2, y: 3},
			},
		},
		{
			name: "correct_distance_lower_y",
			want: 2,
			fields: fields{
				X: 2,
				Y: 5,
			},
			args: args{
				o: &Location{x: 2, y: 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Location{x: tt.fields.X, y: tt.fields.Y}
			if got := l.Distance(tt.args.o); got != tt.want {
				t.Errorf("Distance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocation_Equals(t *testing.T) {
	type fields struct {
		X int
		Y int
	}
	type args struct {
		o *Location
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "correct_equal",
			want: true,
			fields: fields{
				X: 2,
				Y: 3,
			},
			args: args{
				o: &Location{x: 2, y: 3},
			},
		},
		{
			name: "correct_not_equal_x_y",
			want: false,
			fields: fields{
				X: 2,
				Y: 3,
			},
			args: args{
				o: &Location{x: 4, y: 5},
			},
		},
		{
			name: "correct_not_equal_x",
			want: false,
			fields: fields{
				X: 2,
				Y: 3,
			},
			args: args{
				o: &Location{x: 4, y: 3},
			},
		},
		{
			name: "correct_not_equal_y",
			want: false,
			fields: fields{
				X: 2,
				Y: 3,
			},
			args: args{
				o: &Location{x: 2, y: 5},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Location{x: tt.fields.X, y: tt.fields.Y}
			if got := l.Equals(tt.args.o); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLocation(t *testing.T) {
	type args struct {
		x int
		y int
	}
	tests := []struct {
		name    string
		args    args
		want    *Location
		wantErr bool
	}{
		{
			name:    "correct_location",
			want:    &Location{x: 4, y: 3},
			wantErr: false,
			args: args{
				x: 4,
				y: 3,
			},
		},
		{
			name:    "correct_location_min",
			want:    &Location{x: 1, y: 1},
			wantErr: false,
			args: args{
				x: 1,
				y: 1,
			},
		},
		{
			name:    "correct_location_max",
			want:    &Location{x: 10, y: 10},
			wantErr: false,
			args: args{
				x: 10,
				y: 10,
			},
		},
		{
			name:    "incorrect_location_max_x",
			want:    nil,
			wantErr: true,
			args: args{
				x: 11,
				y: 10,
			},
		},
		{
			name:    "incorrect_location_max_y",
			want:    nil,
			wantErr: true,
			args: args{
				x: 10,
				y: 11,
			},
		},
		{
			name:    "incorrect_location_max_x_y",
			want:    nil,
			wantErr: true,
			args: args{
				x: 11,
				y: 11,
			},
		},
		{
			name:    "incorrect_location_min_x",
			want:    nil,
			wantErr: true,
			args: args{
				x: 0,
				y: 1,
			},
		},
		{
			name:    "incorrect_location_max_y",
			want:    nil,
			wantErr: true,
			args: args{
				x: 1,
				y: 0,
			},
		},
		{
			name:    "incorrect_location_max_x_y",
			want:    nil,
			wantErr: true,
			args: args{
				x: 0,
				y: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLocation(tt.args.x, tt.args.y)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLocation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomLocation(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "random_location_creation",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandomLocation(); got == nil {
				t.Errorf("RandomLocation() got nil, want not nil")
			}
		})
	}
}

func TestLocation_IsEmpty(t *testing.T) {
	type fields struct {
		X int
		Y int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "is_empty",
			fields: fields{
				X: 0,
				Y: 0,
			},
			want: true,
		},
		{
			name: "is_not_empty",
			fields: fields{
				X: 1,
				Y: 1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Location{x: tt.fields.X, y: tt.fields.Y}
			if got := l.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
