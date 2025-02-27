package column

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bytehouse-cloud/driver-go/driver/lib/ch_encoding"
)

func TestDateTime64ColumnData_ReadFromTexts(t *testing.T) {
	type args struct {
		texts []string
	}
	tests := []struct {
		name            string
		args            args
		wantDataWritten []string
		wantRowsRead    int
		wantErr         bool
	}{
		{
			name: "Should write data and return number of rows read with no error",
			args: args{
				texts: []string{"1950-01-02 15:04:05.000", "2020-01-02 15:04:05.000", "2019-01-01 00:00:00.000"},
			},
			wantDataWritten: []string{"1950-01-02 15:04:05.000", "2020-01-02 15:04:05.000", "2019-01-01 00:00:00.000"},
			wantRowsRead:    3,
			wantErr:         false,
		},
		{
			name: "Given different format then no error",
			args: args{
				texts: []string{"1950-01-02", "2020-01-02 15:04:05", "2020-01-02 15:04:05.322"},
			},
			wantDataWritten: []string{"1950-01-02 00:00:00.000", "2020-01-02 15:04:05.000", "2020-01-02 15:04:05.322"},
			wantRowsRead:    3,
			wantErr:         false,
		},
		{
			name: "Should write data and return number of rows read with no error, empty string",
			args: args{
				texts: []string{"", "1950-01-02 15:04:05.000", "2020-01-02 15:04:05.000"},
			},
			wantDataWritten: []string{"1970-01-01 00:00:00.000", "1950-01-02 15:04:05.000", "2020-01-02 15:04:05.000"},
			wantRowsRead:    3,
			wantErr:         false,
		},
		{
			name: "Should throw error if invalid time format",
			args: args{
				texts: []string{"1970-01-02 15:04:05", "2020-01-02pp 15:04:05"},
			},
			wantRowsRead: 1,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := MustMakeColumnData("DateTime64(3)", 1000)

			got, err := i.ReadFromTexts(tt.args.texts)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if got != tt.wantRowsRead {
				t.Errorf("ReadFromTexts() got = %v, wantRowsRead %v", got, tt.wantRowsRead)
			}

			if len(tt.wantDataWritten) > 0 {
				for index, value := range tt.wantDataWritten {
					if !tt.wantErr {
						assert.Equal(t, value, i.GetString(index))
					}
				}
				return
			}

			for index, value := range tt.args.texts {
				if !tt.wantErr {
					// Only check if is same date, ignore time value as there may be time zone differences
					assert.Equal(t, value, i.GetString(index))
				}
			}
		})
	}
}

func TestDateTime64ColumnData_ReadFromValues(t *testing.T) {
	type args struct {
		values []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Should return the same time value",
			args: args{
				values: []interface{}{
					time.Unix(0, -900000000000000000), time.Now(),
				},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "Should throw error if one of the values is not time.Time",
			args: args{
				values: []interface{}{
					time.Now(), 123,
				},
			},
			want:    1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := MustMakeColumnData("DateTime64(6)", 1000)
			got, err := d.ReadFromValues(tt.args.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFromValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadFromValues() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateTime64ColumnData_EncoderDecoder(t *testing.T) {
	type args struct {
		texts []string
	}
	tests := []struct {
		name            string
		args            args
		wantDataWritten []string
		wantRowsRead    int
		wantErr         bool
	}{
		{
			name: "Should write data and return number of rows read with no error",
			args: args{
				texts: []string{"1950-01-02 15:04:05.000", "2020-01-02 15:04:05.000", "2019-01-01 00:00:00.000"},
			},
			wantRowsRead: 3,
			wantErr:      false,
		},
		{
			name: "Given different format then no error",
			args: args{
				texts: []string{"1950-01-02", "2020-01-02 15:04:05", "2020-01-02 15:04:05.322"},
			},
			wantDataWritten: []string{"1950-01-02 00:00:00.000", "2020-01-02 15:04:05.000", "2020-01-02 15:04:05.322"},
			wantRowsRead:    3,
			wantErr:         false,
		},
		{
			name: "Should write data and return number of rows read with no error, empty string",
			args: args{
				texts: []string{"", "1950-01-02 15:04:05.000", "2020-01-02 15:04:05.000"},
			},
			wantDataWritten: []string{"1970-01-01 00:00:00.000", "1950-01-02 15:04:05.000", "2020-01-02 15:04:05.000"},
			wantRowsRead:    3,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			encoder := ch_encoding.NewEncoder(&buffer)
			decoder := ch_encoding.NewDecoder(&buffer)

			// Write to encoder
			original := MustMakeColumnData("DateTime64(3)", len(tt.args.texts))
			got, err := original.ReadFromTexts(tt.args.texts)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, got, tt.wantRowsRead)
			require.NoError(t, err)
			err = original.WriteToEncoder(encoder)
			require.NoError(t, err)

			// Read from decoder
			newCopy := MustMakeColumnData("DateTime64(3)", len(tt.args.texts))
			err = newCopy.ReadFromDecoder(decoder)

			for index, value := range tt.wantDataWritten {
				if !tt.wantErr {
					require.Equal(t, value, newCopy.GetString(index))
				}
			}

			require.Equal(t, newCopy.Len(), original.Len())
			require.Equal(t, newCopy.Zero(), original.Zero())
			require.Equal(t, newCopy.ZeroString(), original.ZeroString())
			require.NoError(t, original.Close())
			require.NoError(t, newCopy.Close())
		})
	}
}
