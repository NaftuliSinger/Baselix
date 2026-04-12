package types

import (
	"testing"
	"time"
)

func makeRecord(id string, fields map[string]any) RecordResponse {
	r := RecordResponse{ID: id, CreatedAt: time.Time{}, UpdatedAt: time.Time{}}
	for k, v := range fields {
		r.Values = append(r.Values, RecordField{Key: k, Value: v})
	}
	return r
}

func makeRecordWithTimes(id string, createdAt, updatedAt time.Time, fields map[string]any) RecordResponse {
	r := RecordResponse{ID: id, CreatedAt: createdAt, UpdatedAt: updatedAt}
	for k, v := range fields {
		r.Values = append(r.Values, RecordField{Key: k, Value: v})
	}
	return r
}

// fieldValue returns the value for a key from a RecordResponse, or nil if absent.
func fieldValue(r RecordResponse, key string) any {
	for _, f := range r.Values {
		if f.Key == key {
			return f.Value
		}
	}
	return nil
}

func TestApplySorts_EmptyRecords(t *testing.T) {
	result, err := ApplySorts(nil, []Sort{{Field: "name", Dir: "asc"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d records", len(result))
	}
}

func TestApplySorts_EmptySorts(t *testing.T) {
	records := []RecordResponse{
		makeRecord("1", map[string]any{"name": "Charlie"}),
		makeRecord("2", map[string]any{"name": "Alice"}),
	}
	result, err := ApplySorts(records, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].ID != "1" || result[1].ID != "2" {
		t.Errorf("records should be unchanged, got IDs %s, %s", result[0].ID, result[1].ID)
	}
}

func TestApplySorts_FieldNotFound(t *testing.T) {
	records := []RecordResponse{
		makeRecord("1", map[string]any{"name": "Alice"}),
	}
	_, err := ApplySorts(records, []Sort{{Field: "missing", Dir: "asc"}})
	if err == nil {
		t.Error("expected error for missing sort field, got nil")
	}
}

func TestApplySorts_SingleSortAsc(t *testing.T) {
	records := []RecordResponse{
		makeRecord("1", map[string]any{"name": "Charlie"}),
		makeRecord("2", map[string]any{"name": "Alice"}),
		makeRecord("3", map[string]any{"name": "Bob"}),
	}
	result, err := ApplySorts(records, []Sort{{Field: "name", Dir: "asc"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"Alice", "Bob", "Charlie"}
	for i, w := range want {
		if got := fieldValue(result[i], "name"); got != w {
			t.Errorf("index %d: expected %q, got %v", i, w, got)
		}
	}
}

func TestApplySorts_SingleSortDesc(t *testing.T) {
	records := []RecordResponse{
		makeRecord("1", map[string]any{"age": float64(30)}),
		makeRecord("2", map[string]any{"age": float64(25)}),
		makeRecord("3", map[string]any{"age": float64(40)}),
	}
	result, err := ApplySorts(records, []Sort{{Field: "age", Dir: "desc"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []float64{40, 30, 25}
	for i, w := range want {
		if got := fieldValue(result[i], "age"); got != w {
			t.Errorf("index %d: expected %v, got %v", i, w, got)
		}
	}
}

func TestApplySorts_MultiSort(t *testing.T) {
	records := []RecordResponse{
		makeRecord("1", map[string]any{"dept": "Engineering", "name": "Charlie"}),
		makeRecord("2", map[string]any{"dept": "Engineering", "name": "Alice"}),
		makeRecord("3", map[string]any{"dept": "Marketing", "name": "Bob"}),
		makeRecord("4", map[string]any{"dept": "Engineering", "name": "Bob"}),
	}
	// Sort by dept asc, then name asc
	result, err := ApplySorts(records, []Sort{
		{Field: "dept", Dir: "asc"},
		{Field: "name", Dir: "asc"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantIDs := []string{"2", "4", "1", "3"} // Engineering:Alice, Engineering:Bob, Engineering:Charlie, Marketing:Bob
	for i, id := range wantIDs {
		if result[i].ID != id {
			t.Errorf("index %d: expected ID %q, got %q (name=%v)", i, id, result[i].ID, fieldValue(result[i], "name"))
		}
	}
}

func TestApplySorts_SortDirCaseInsensitive(t *testing.T) {
	records := []RecordResponse{
		makeRecord("1", map[string]any{"score": float64(10)}),
		makeRecord("2", map[string]any{"score": float64(50)}),
		makeRecord("3", map[string]any{"score": float64(30)}),
	}
	result, err := ApplySorts(records, []Sort{{Field: "score", Dir: "DESC"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []float64{50, 30, 10}
	for i, w := range want {
		if got := fieldValue(result[i], "score"); got != w {
			t.Errorf("index %d: expected %v, got %v", i, w, got)
		}
	}
}

func TestApplySorts_SortByID(t *testing.T) {
	records := []RecordResponse{
		makeRecord("charlie", nil),
		makeRecord("alice", nil),
		makeRecord("bob", nil),
	}
	result, err := ApplySorts(records, []Sort{{Field: "id", Dir: "asc"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantIDs := []string{"alice", "bob", "charlie"}
	for i, id := range wantIDs {
		if result[i].ID != id {
			t.Errorf("index %d: expected ID %q, got %q", i, id, result[i].ID)
		}
	}
}

func TestApplySorts_SortByCreatedAt(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	records := []RecordResponse{
		makeRecordWithTimes("1", base.Add(2*time.Hour), time.Time{}, nil),
		makeRecordWithTimes("2", base, time.Time{}, nil),
		makeRecordWithTimes("3", base.Add(time.Hour), time.Time{}, nil),
	}
	result, err := ApplySorts(records, []Sort{{Field: "created_at", Dir: "asc"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantIDs := []string{"2", "3", "1"}
	for i, id := range wantIDs {
		if result[i].ID != id {
			t.Errorf("index %d: expected ID %q, got %q", i, id, result[i].ID)
		}
	}
}

func TestApplySorts_SortByUpdatedAtDesc(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	records := []RecordResponse{
		makeRecordWithTimes("1", time.Time{}, base.Add(time.Hour), nil),
		makeRecordWithTimes("2", time.Time{}, base.Add(3*time.Hour), nil),
		makeRecordWithTimes("3", time.Time{}, base.Add(2*time.Hour), nil),
	}
	result, err := ApplySorts(records, []Sort{{Field: "updated_at", Dir: "desc"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantIDs := []string{"2", "3", "1"}
	for i, id := range wantIDs {
		if result[i].ID != id {
			t.Errorf("index %d: expected ID %q, got %q", i, id, result[i].ID)
		}
	}
}

func TestApplySorts_MetaFieldsNotAffectedByMissingCheck(t *testing.T) {
	// Sorting by meta fields should never return a "field does not exist" error,
	// even when records have no Values entries.
	records := []RecordResponse{
		makeRecord("1", nil),
		makeRecord("2", nil),
	}
	for _, field := range []string{"id", "created_at", "updated_at"} {
		_, err := ApplySorts(records, []Sort{{Field: field, Dir: "asc"}})
		if err != nil {
			t.Errorf("meta field %q: unexpected error: %v", field, err)
		}
	}
}
