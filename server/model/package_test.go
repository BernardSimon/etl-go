package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomTime_MarshalJSON(t *testing.T) {
	ct := &CustomTime{Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.Local)}
	data, err := ct.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, `"2024-01-15 10:30:00"`, string(data))
}

func TestCustomTime_MarshalJSON_Zero(t *testing.T) {
	ct := &CustomTime{}
	data, err := ct.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, "0000-00-00 00:00:00", string(data))
}

func TestCustomTime_MarshalJSON_Nil(t *testing.T) {
	var ct *CustomTime
	data, err := ct.MarshalJSON()
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestCustomTime_UnmarshalJSON_DateTime(t *testing.T) {
	ct := &CustomTime{}
	err := ct.UnmarshalJSON([]byte(`"2024-01-15 10:30:00"`))
	require.NoError(t, err)
	assert.Equal(t, 2024, ct.Time.Year())
	assert.Equal(t, time.Month(1), ct.Time.Month())
	assert.Equal(t, 15, ct.Time.Day())
}

func TestCustomTime_UnmarshalJSON_Date(t *testing.T) {
	ct := &CustomTime{}
	err := ct.UnmarshalJSON([]byte(`"2024-01-15"`))
	require.NoError(t, err)
	assert.Equal(t, 2024, ct.Time.Year())
}

func TestCustomTime_UnmarshalJSON_Timestamp(t *testing.T) {
	ct := &CustomTime{}
	// 1705312200000 is roughly 2024-01-15 in millis
	err := ct.UnmarshalJSON([]byte(`"1705312200000"`))
	require.NoError(t, err)
	assert.Equal(t, 2024, ct.Time.Year())
}

func TestCustomTime_UnmarshalJSON_Empty(t *testing.T) {
	ct := &CustomTime{}
	err := ct.UnmarshalJSON([]byte(`""`))
	require.NoError(t, err)
	assert.True(t, ct.Time.IsZero())
}

func TestCustomTime_Scan_TimeValue(t *testing.T) {
	ct := &CustomTime{}
	now := time.Now()
	err := ct.Scan(now)
	require.NoError(t, err)
	assert.Equal(t, now, ct.Time)
}

func TestCustomTime_Scan_Nil(t *testing.T) {
	ct := &CustomTime{}
	err := ct.Scan(nil)
	require.NoError(t, err)
	assert.True(t, ct.Time.IsZero())
}

func TestCustomTime_Value(t *testing.T) {
	now := time.Now()
	ct := &CustomTime{Time: now}
	val, err := ct.Value()
	require.NoError(t, err)
	assert.Equal(t, now, val)
}

func TestCustomTime_Value_Zero(t *testing.T) {
	ct := &CustomTime{}
	val, err := ct.Value()
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestModel_BeforeCreate_GeneratesULID(t *testing.T) {
	m := &Model{}
	err := m.BeforeCreate(nil)
	require.NoError(t, err)
	assert.NotEmpty(t, m.ID)
	assert.Len(t, m.ID, 26) // ULID format
}

func TestModel_BeforeCreate_PreservesExistingID(t *testing.T) {
	m := &Model{ID: "existing-id"}
	err := m.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, "existing-id", m.ID)
}
