package templates

import (
	"reflect"
	"testing"
)

func TestParseKV(t *testing.T) {
	type args struct {
		kvStr string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Test empty string",
			args: args{
				kvStr: "",
			},
			want: map[string]string{},
		},
		{
			name: "Test malformed string",
			args: args{
				kvStr: "keyword1",
			},
			want: map[string]string{},
		},
		{
			name: "Test key with no value",
			args: args{
				kvStr: "keyword1=",
			},
			want: map[string]string{"keyword1": ""},
		},
		{
			name: "TestParseKV 1",
			args: args{
				kvStr: "keyword1=value1,keyword2=value2,keyword3=value3,value4,value5,keyword4=value6",
			},
			want: map[string]string{
				"keyword1": "value1",
				"keyword2": "value2",
				"keyword3": "value3,value4,value5",
				"keyword4": "value6",
			},
		},
		{
			name: "TestParseKV 2",
			args: args{
				kvStr: "keyword1=value1,keyword2=value2,keyword3=value3,value4,value5:value6,value7,value8,keyword4=value9",
			},
			want: map[string]string{
				"keyword1": "value1",
				"keyword2": "value2",
				"keyword3": "value3,value4,value5:value6,value7,value8",
				"keyword4": "value9",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseKV(tt.args.kvStr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseKV() = %v, want %v", got, tt.want)
			}
		})
	}
}
