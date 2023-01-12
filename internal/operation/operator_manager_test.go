package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-to-k/delstack/internal/io"
)

/*
	Test Cases
*/

func TestOperatorManager_getOperatorResourcesLength(t *testing.T) {
	io.NewLogger(false)

	mock := NewMockOperatorCollection()

	type args struct {
		ctx  context.Context
		mock IOperatorCollection
	}

	cases := []struct {
		name string
		args args
		want int
	}{
		{
			name: "get operator resources length successfully",
			args: args{
				ctx:  context.Background(),
				mock: mock,
			},
			want: 6,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			operatorManager := NewOperatorManager(tt.args.mock)

			got := operatorManager.getOperatorResourcesLength()
			if got != tt.want {
				t.Errorf("got = %#v, want %#v", got, tt.want)
				return
			}
		})
	}
}

func TestOperatorManager_CheckResourceCounts(t *testing.T) {
	io.NewLogger(false)

	mock := NewMockOperatorCollection()
	incorrectResourceCountsMock := NewIncorrectResourceCountsMockOperatorCollection()

	type args struct {
		ctx  context.Context
		mock IOperatorCollection
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "check resource counts successfully",
			args: args{
				ctx:  context.Background(),
				mock: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "check resource counts failure",
			args: args{
				ctx:  context.Background(),
				mock: incorrectResourceCountsMock,
			},
			want:    fmt.Errorf("UnsupportedResourceError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			operatorManager := NewOperatorManager(tt.args.mock)

			err := operatorManager.CheckResourceCounts()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestOperatorManager_DeleteResourceCollection(t *testing.T) {
	io.NewLogger(false)

	mock := NewMockOperatorCollection()
	operatorDeleteResourcesMock := NewOperatorDeleteResourcesMockOperatorCollection()

	type args struct {
		ctx  context.Context
		mock IOperatorCollection
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete resource collection successfully",
			args: args{
				ctx:  context.Background(),
				mock: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resource collection failure",
			args: args{
				ctx:  context.Background(),
				mock: operatorDeleteResourcesMock,
			},
			want:    fmt.Errorf("ErrorDeleteResources"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			operatorManager := NewOperatorManager(tt.args.mock)

			err := operatorManager.DeleteResourceCollection(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
