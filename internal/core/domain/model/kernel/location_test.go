package kernel

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDefaultLocation(t *testing.T) {
	tests := []struct {
		name    string
		want    Location
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			want: Location{
				x: 1,
				y: 1,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDefaultLocation()

			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want.x, got.X())
			assert.Equal(t, tt.want.y, got.Y())
			assert.Equal(t, fmt.Sprintf("(%d, %d)", tt.want.x, tt.want.y), got.String())
		})
	}
}

func TestNewLocation(t *testing.T) {
	type args struct {
		x int8
		y int8
	}
	tests := []struct {
		name    string
		args    args
		want    Location
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "success",
			args:    args{3, 8},
			want:    Location{x: 3, y: 8},
			wantErr: assert.NoError,
		},
		{
			name:    "success min",
			args:    args{1, 1},
			want:    Location{x: 1, y: 1},
			wantErr: assert.NoError,
		},
		{
			name:    "success max",
			args:    args{10, 10},
			want:    Location{x: 10, y: 10},
			wantErr: assert.NoError,
		},
		{
			name:    "invalid x min",
			args:    args{0, 1},
			want:    invalidLocation,
			wantErr: invalidLocationCheck,
		},
		{
			name:    "invalid y min",
			args:    args{1, 0},
			want:    invalidLocation,
			wantErr: invalidLocationCheck,
		},
		{
			name:    "invalid x and y min",
			args:    args{0, 0},
			want:    invalidLocation,
			wantErr: invalidLocationCheck,
		},
		{
			name:    "invalid x max",
			args:    args{11, 10},
			want:    invalidLocation,
			wantErr: invalidLocationCheck,
		},
		{
			name:    "invalid y max",
			args:    args{10, 11},
			want:    invalidLocation,
			wantErr: invalidLocationCheck,
		},
		{
			name:    "invalid x and y max",
			args:    args{11, 11},
			want:    invalidLocation,
			wantErr: invalidLocationCheck,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLocation(tt.args.x, tt.args.y)

			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want.x, got.X())
			assert.Equal(t, tt.want.y, got.Y())
			assert.Equal(t, fmt.Sprintf("(%d, %d)", tt.want.x, tt.want.y), got.String())
		})
	}
}

func TestLocation_Distance(t *testing.T) {
	type fields struct {
		x int8
		y int8
	}
	type args struct {
		other Location
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int8
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			fields: fields{
				x: 2,
				y: 6,
			},
			args: args{
				other: Location{
					x: 4,
					y: 9,
				},
			},
			want:    5,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := NewLocation(tt.fields.x, tt.fields.y)

			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, l.DistanceTo(tt.args.other))
		})
	}
}

func TestLocation_Equals(t *testing.T) {
	type fields struct {
		x int8
		y int8
	}
	type args struct {
		other Location
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			fields: fields{
				x: 2,
				y: 6,
			},
			args: args{
				other: Location{
					x: 2,
					y: 6,
				},
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "failed x",
			fields: fields{
				x: 2,
				y: 6,
			},
			args: args{
				other: Location{
					x: 3,
					y: 6,
				},
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "failed y",
			fields: fields{
				x: 2,
				y: 6,
			},
			args: args{
				other: Location{
					x: 2,
					y: 7,
				},
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "failed x & y",
			fields: fields{
				x: 2,
				y: 6,
			},
			args: args{
				other: Location{
					x: 3,
					y: 7,
				},
			},
			want:    false,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := NewLocation(tt.fields.x, tt.fields.y)

			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, l.Equals(tt.args.other))
		})
	}
}

func invalidLocationCheck(t assert.TestingT, err error, _ ...interface{}) bool {
	return assert.Error(t, err) && assert.True(t, errors.Is(err, InvalidLocationErr))
}
