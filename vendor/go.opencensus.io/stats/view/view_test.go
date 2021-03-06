// Copyright 2017, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package view

import (
	"context"
	"testing"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

func Test_View_MeasureFloat64_AggregationDistribution(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")
	k3, _ := tag.NewKey("k3")
	agg1 := Distribution(2)
	m, _ := stats.Int64("Test_View_MeasureFloat64_AggregationDistribution/m1", "", stats.UnitNone)
	view1 := &View{
		TagKeys:     []tag.Key{k1, k2},
		Measure:     m,
		Aggregation: agg1,
	}
	view, err := newViewInternal(view1)
	if err != nil {
		t.Fatal(err)
	}

	type tagString struct {
		k tag.Key
		v string
	}
	type record struct {
		f    float64
		tags []tagString
	}

	type testCase struct {
		label    string
		records  []record
		wantRows []*Row
	}

	tcs := []testCase{
		{
			"1",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					&DistributionData{
						2, 1, 5, 3, 8, []int64{1, 1}, []float64{2},
					},
				},
			},
		},
		{
			"2",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k2, "v2"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					&DistributionData{
						1, 1, 1, 1, 0, []int64{1, 0}, []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					&DistributionData{
						1, 5, 5, 5, 0, []int64{0, 1}, []float64{2},
					},
				},
			},
		},
		{
			"3",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}, {k3, "v3"}}},
				{1, []tagString{{k1, "v1 other"}}},
				{5, []tagString{{k2, "v2"}}},
				{5, []tagString{{k1, "v1"}, {k2, "v2"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					&DistributionData{
						2, 1, 5, 3, 8, []int64{1, 1}, []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1 other"}},
					&DistributionData{
						1, 1, 1, 1, 0, []int64{1, 0}, []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					&DistributionData{
						1, 5, 5, 5, 0, []int64{0, 1}, []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1"}, {Key: k2, Value: "v2"}},
					&DistributionData{
						1, 5, 5, 5, 0, []int64{0, 1}, []float64{2},
					},
				},
			},
		},
		{
			"4",
			[]record{
				{1, []tagString{{k1, "v1 is a very long value key"}}},
				{5, []tagString{{k1, "v1 is a very long value key"}, {k3, "v3"}}},
				{1, []tagString{{k1, "v1 is another very long value key"}}},
				{1, []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
				{5, []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
				{3, []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
				{3, []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1 is a very long value key"}},
					&DistributionData{
						2, 1, 5, 3, 8, []int64{1, 1}, []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1 is another very long value key"}},
					&DistributionData{
						1, 1, 1, 1, 0, []int64{1, 0}, []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1 is a very long value key"}, {Key: k2, Value: "v2 is a very long value key"}},
					&DistributionData{
						4, 1, 5, 3, 2.66666666666667 * 3, []int64{1, 3}, []float64{2},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		view.clearRows()
		view.subscribe()
		for _, r := range tc.records {
			mods := []tag.Mutator{}
			for _, t := range r.tags {
				mods = append(mods, tag.Insert(t.k, t.v))
			}
			ctx, err := tag.New(context.Background(), mods...)
			if err != nil {
				t.Errorf("%v: NewMap = %v", tc.label, err)
			}
			view.addSample(tag.FromContext(ctx), r.f)
		}

		gotRows := view.collectedRows()
		for i, got := range gotRows {
			if !containsRow(tc.wantRows, got) {
				t.Errorf("%v-%d: got row %v; want none", tc.label, i, got)
				break
			}
		}

		for i, want := range tc.wantRows {
			if !containsRow(gotRows, want) {
				t.Errorf("%v-%d: got none; want row %v", tc.label, i, want)
				break
			}
		}
	}
}

func Test_View_MeasureFloat64_AggregationSum(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")
	k3, _ := tag.NewKey("k3")
	m, _ := stats.Int64("Test_View_MeasureFloat64_AggregationSum/m1", "", stats.UnitNone)
	view, err := newViewInternal(&View{TagKeys: []tag.Key{k1, k2}, Measure: m, Aggregation: Sum()})
	if err != nil {
		t.Fatal(err)
	}

	type tagString struct {
		k tag.Key
		v string
	}
	type record struct {
		f    float64
		tags []tagString
	}

	tcs := []struct {
		label    string
		records  []record
		wantRows []*Row
	}{
		{
			"1",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					newSumData(6),
				},
			},
		},
		{
			"2",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k2, "v2"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					newSumData(1),
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					newSumData(5),
				},
			},
		},
		{
			"3",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}, {k3, "v3"}}},
				{1, []tagString{{k1, "v1 other"}}},
				{5, []tagString{{k2, "v2"}}},
				{5, []tagString{{k1, "v1"}, {k2, "v2"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					newSumData(6),
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1 other"}},
					newSumData(1),
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					newSumData(5),
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1"}, {Key: k2, Value: "v2"}},
					newSumData(5),
				},
			},
		},
	}

	for _, tt := range tcs {
		view.clearRows()
		view.subscribe()
		for _, r := range tt.records {
			mods := []tag.Mutator{}
			for _, t := range r.tags {
				mods = append(mods, tag.Insert(t.k, t.v))
			}
			ctx, err := tag.New(context.Background(), mods...)
			if err != nil {
				t.Errorf("%v: New = %v", tt.label, err)
			}
			view.addSample(tag.FromContext(ctx), r.f)
		}

		gotRows := view.collectedRows()
		for i, got := range gotRows {
			if !containsRow(tt.wantRows, got) {
				t.Errorf("%v-%d: got row %v; want none", tt.label, i, got)
				break
			}
		}

		for i, want := range tt.wantRows {
			if !containsRow(gotRows, want) {
				t.Errorf("%v-%d: got none; want row %v", tt.label, i, want)
				break
			}
		}
	}
}

func TestCanonicalize(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")
	m, _ := stats.Int64("TestCanonicalize/m1", "desc desc", stats.UnitNone)
	v := &View{TagKeys: []tag.Key{k2, k1}, Measure: m, Aggregation: Mean()}
	err := v.canonicalize()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := v.Name, "TestCanonicalize/m1"; got != want {
		t.Errorf("vc.Name = %q; want %q", got, want)
	}
	if got, want := v.Description, "desc desc"; got != want {
		t.Errorf("vc.Description = %q; want %q", got, want)
	}
	if got, want := len(v.TagKeys), 2; got != want {
		t.Errorf("len(vc.TagKeys) = %d; want %d", got, want)
	}
	if got, want := v.TagKeys[0].Name(), "k1"; got != want {
		t.Errorf("vc.TagKeys[0].Name() = %q; want %q", got, want)
	}
}

func Test_View_MeasureFloat64_AggregationMean(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")
	k3, _ := tag.NewKey("k3")
	m, _ := stats.Int64("Test_View_MeasureFloat64_AggregationMean/m1", "", stats.UnitNone)
	viewDesc := &View{TagKeys: []tag.Key{k1, k2}, Measure: m, Aggregation: Mean()}
	view, err := newViewInternal(viewDesc)
	if err != nil {
		t.Fatal(err)
	}

	type tagString struct {
		k tag.Key
		v string
	}
	type record struct {
		f    float64
		tags []tagString
	}

	tcs := []struct {
		label    string
		records  []record
		wantRows []*Row
	}{
		{
			"1",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					newMeanData(3, 2),
				},
			},
		},
		{
			"2",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k2, "v2"}}},
				{-0.5, []tagString{{k2, "v2"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					newMeanData(1, 1),
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					newMeanData(2.25, 2),
				},
			},
		},
		{
			"3",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}, {k3, "v3"}}},
				{1, []tagString{{k1, "v1 other"}}},
				{5, []tagString{{k2, "v2"}}},
				{5, []tagString{{k1, "v1"}, {k2, "v2"}}},
				{-4, []tagString{{k1, "v1"}, {k2, "v2"}}},
			},
			[]*Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					newMeanData(3, 2),
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1 other"}},
					newMeanData(1, 1),
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					newMeanData(5, 1),
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1"}, {Key: k2, Value: "v2"}},
					newMeanData(0.5, 2),
				},
			},
		},
	}

	for _, tt := range tcs {
		view.clearRows()
		view.subscribe()
		for _, r := range tt.records {
			mods := []tag.Mutator{}
			for _, t := range r.tags {
				mods = append(mods, tag.Insert(t.k, t.v))
			}
			ctx, err := tag.New(context.Background(), mods...)
			if err != nil {
				t.Errorf("%v: New = %v", tt.label, err)
			}
			view.addSample(tag.FromContext(ctx), r.f)
		}

		gotRows := view.collectedRows()
		for i, got := range gotRows {
			if !containsRow(tt.wantRows, got) {
				t.Errorf("%v-%d: got row %v; want none", tt.label, i, got)
				break
			}
		}

		for i, want := range tt.wantRows {
			if !containsRow(gotRows, want) {
				t.Errorf("%v-%d: got none; want row %v", tt.label, i, want)
				break
			}
		}
	}
}

func TestViewSortedKeys(t *testing.T) {
	k1, _ := tag.NewKey("a")
	k2, _ := tag.NewKey("b")
	k3, _ := tag.NewKey("c")
	ks := []tag.Key{k1, k3, k2}

	m, _ := stats.Int64("TestViewSortedKeys/m1", "", stats.UnitNone)
	Subscribe(&View{
		Name:        "sort_keys",
		Description: "desc sort_keys",
		TagKeys:     ks,
		Measure:     m,
		Aggregation: Mean(),
	})
	// Subscribe normalizes the view by sorting the tag keys, retrieve the normalized view
	v := Find("sort_keys")

	want := []string{"a", "b", "c"}
	vks := v.TagKeys
	if len(vks) != len(want) {
		t.Errorf("Keys = %+v; want %+v", vks, want)
	}

	for i, v := range want {
		if got, want := v, vks[i].Name(); got != want {
			t.Errorf("View name = %q; want %q", got, want)
		}
	}
}

// containsRow returns true if rows contain r.
func containsRow(rows []*Row, r *Row) bool {
	for _, x := range rows {
		if r.Equal(x) {
			return true
		}
	}
	return false
}
