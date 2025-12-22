package copyutil

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

type User struct {
	Age   int
	Birth time.Time
	Name  string
}

type Engineer struct {
	User  *User
	Title string
	Args  map[string]string
}

type Good struct {
	Name string
	Age  int
}

type Flower struct {
	Title string
	Args  string
	User  *Good
}

func TestClone(t *testing.T) {
	type args struct {
		source any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{name: "user", args: args{source: &User{Name: "abc", Age: 20}}, want: &User{Name: "abc", Age: 20}, wantErr: false},
		{name: "deep", args: args{source: Engineer{User: &User{Name: "abc", Age: 20}, Title: "s"}}, want: Engineer{User: &User{Name: "abc", Age: 20}, Title: "s"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Clone(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Clone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Clone() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	source := &User{Name: "abc", Age: 20}
	target := &Good{}
	err := Copy(source, target)
	assert.Nil(t, err)
	assert.Equal(t, source.Name, target.Name)
	assert.Equal(t, source.Age, target.Age)
	source1 := &Engineer{User: &User{Name: "abc", Age: 20}, Title: "s", Args: map[string]string{"name": "test"}}
	target1 := &Flower{}
	err1 := Copy(source1, target1)
	assert.Nil(t, err1)
	assert.Equal(t, source1.Title, target1.Title)         // copy
	assert.Equal(t, target1.Args, "")                     // type diff, not copy
	assert.Equal(t, source1.User.Name, target1.User.Name) // deep copy
	assert.Equal(t, source1.User.Age, target1.User.Age)   // deep copy
}

func TestCopyNoNil(t *testing.T) {
	source1 := &Engineer{User: nil, Title: "s", Args: nil}
	target1 := &Flower{User: &Good{Name: "ab"}}
	err1 := CopyNoNil(source1, target1)
	assert.Nil(t, err1)
	assert.Equal(t, source1.Title, target1.Title) // copy
	assert.Equal(t, target1.Args, "")             // type diff, not copy
	assert.Equal(t, "ab", target1.User.Name)      // nil not copy
	assert.Equal(t, 0, target1.User.Age)          // nil not copy
}

func TestConvert(t *testing.T) {
	source1 := &Engineer{User: &User{Name: "abc", Age: 20}, Title: "s", Args: nil}
	target1 := &Flower{Args: "args"}
	err1 := Convert(source1, target1)
	assert.Nil(t, err1)
	assert.Equal(t, source1.Title, target1.Title)         // copy
	assert.Equal(t, target1.Args, "args")                 // type diff, not copy
	assert.Equal(t, source1.User.Name, target1.User.Name) // deep copy
	assert.Equal(t, source1.User.Age, target1.User.Age)   // deep copy
}

func TestConvertToStruct(t *testing.T) {
	source := &User{Name: "abc", Age: 20}
	structData, err := ConvertToStruct(source)
	assert.Nil(t, err)
	structDataJson, _ := json.Marshal(structData)
	sourceDataJson, _ := json.Marshal(source)
	assert.Equal(t, string(structDataJson), string(sourceDataJson))
}
