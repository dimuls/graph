package dijkstra

import (
	"reflect"
	"testing"

	"github.com/dimuls/graph/entity"
)

func TestShortestPath(t *testing.T) {
	type args struct {
		vs   []entity.Vertex
		es   []entity.Edge
		from int64
		to   int64
	}
	tests := []struct {
		name    string
		args    args
		want    []int64
		wantErr bool
	}{
		{
			name: "one vertex",
			args: args{
				vs:   []entity.Vertex{{}},
				es:   nil,
				from: 0,
				to:   0,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "one vertex with cycle edge",
			args: args{
				vs: []entity.Vertex{{}},
				es: []entity.Edge{{
					ID:     0,
					From:   0,
					To:     0,
					Weight: 1,
				}},
				from: 0,
				to:   0,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "two vertexes one edge",
			args: args{
				vs: []entity.Vertex{{}, {ID: 1}},
				es: []entity.Edge{{
					ID:     0,
					From:   0,
					To:     1,
					Weight: 1,
				}},
				from: 0,
				to:   1,
			},
			want:    []int64{0},
			wantErr: false,
		},
		{
			name: "two vertexes without edge",
			args: args{
				vs:   []entity.Vertex{{}, {ID: 1}},
				es:   nil,
				from: 0,
				to:   1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "two vertexes two edges",
			args: args{
				vs: []entity.Vertex{{}, {ID: 1}},
				es: []entity.Edge{{
					ID:     0,
					From:   0,
					To:     1,
					Weight: 1,
				}, {
					ID:     1,
					From:   1,
					To:     0,
					Weight: 1,
				}},
				from: 0,
				to:   1,
			},
			want:    []int64{0},
			wantErr: false,
		},
		{
			name: "three vertexes three edges",
			args: args{
				vs: []entity.Vertex{{}, {ID: 1}, {ID: 2}},
				es: []entity.Edge{{
					ID:     0,
					From:   0,
					To:     1,
					Weight: 0.5,
				}, {
					ID:     1,
					From:   1,
					To:     2,
					Weight: 0.5,
				}, {
					ID:     2,
					From:   0,
					To:     2,
					Weight: 1.1,
				}},
				from: 0,
				to:   2,
			},
			want:    []int64{0, 1},
			wantErr: false,
		},
		{
			name: "three vertexes three edges 2",
			args: args{
				vs: []entity.Vertex{{}, {ID: 1}, {ID: 2}},
				es: []entity.Edge{{
					ID:     0,
					From:   0,
					To:     1,
					Weight: 10,
				}, {
					ID:     1,
					From:   1,
					To:     2,
					Weight: 11,
				}, {
					ID:     2,
					From:   0,
					To:     2,
					Weight: 15,
				}},
				from: 0,
				to:   2,
			},
			want:    []int64{2},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ShortestPath(tt.args.vs, tt.args.es, tt.args.from, tt.args.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShortestPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShortestPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkShortestPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ShortestPath([]entity.Vertex{{}, {ID: 1}, {ID: 2}},
			[]entity.Edge{{
				ID:     0,
				From:   0,
				To:     1,
				Weight: 0.5,
			}, {
				ID:     1,
				From:   1,
				To:     2,
				Weight: 0.5,
			}, {
				ID:     2,
				From:   0,
				To:     2,
				Weight: 1.1,
			}}, 0, 2)
	}
}
