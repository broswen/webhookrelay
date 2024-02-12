package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNullableTime_MarshalJSON(t *testing.T) {
	now := time.Now()
	tests := []struct {
		Desc     string
		Time     NullableTime
		Expected []byte
		Error    error
	}{
		{
			Desc: "invalid time should be null",
			Time: NullableTime{
				sql.NullTime{
					Time:  time.Time{},
					Valid: false,
				},
			},
			Expected: []byte("null"),
		},
		{
			Desc: "valid time should not be null",
			Time: NullableTime{
				sql.NullTime{
					Time:  now,
					Valid: true,
				},
			},
			Expected: []byte(fmt.Sprintf("\"%s\"", now.UTC().Format(time.RFC3339Nano))),
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, func(t *testing.T) {
			actual, err := json.Marshal(tc.Time)
			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestNullableTime_UnmarshalJSON(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		Desc     string
		Time     string
		Expected NullableTime
		Error    error
	}{
		{
			Desc: "invalid time should throw an error",
			Time: "blah blah",
			Expected: NullableTime{
				sql.NullTime{
					Valid: false,
				},
			},
			Error: &json.SyntaxError{},
		},
		{
			Desc: "valid time should not be null",
			Time: fmt.Sprintf("\"%s\"", now.UTC().Format(time.RFC3339Nano)),
			Expected: NullableTime{
				sql.NullTime{
					Time:  now,
					Valid: true,
				},
			},
		},
		{
			Desc: "null time should be null",
			Time: "null",
			Expected: NullableTime{
				sql.NullTime{
					Valid: false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, func(t *testing.T) {
			actual := NullableTime{}
			err := json.Unmarshal([]byte(tc.Time), &actual)
			if tc.Error != nil {
				assert.ErrorAs(t, err, &tc.Error)
			} else {
				assert.Equal(t, tc.Expected, actual)
			}
		})
	}
}
