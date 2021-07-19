package main

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

func Test_parseXML(t *testing.T) {
	type args struct {
		data []byte
		out  *Products
	}
	var products Products
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"someName",
			args{getData("dataProduct.xml"), &products},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseXML(tt.args.data, tt.args.out); (err != nil) != tt.wantErr {
				t.Errorf("parseXML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func getData(filename string) []byte {

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	b, err := ioutil.ReadAll(file)
	return b
}

func Test_getCurrentUserName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Can get the current user name",
			want: "laszlobogacsi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCurrentUserName(); got != tt.want {
				t.Errorf("getCurrentUserName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filePathsForCurrentUser(t *testing.T) {
	type args struct {
		username string
		files    []string
	}
	tests := []struct {
		name      string
		args      args
		wantPaths []string
	}{
		{
			name:      "Can generate paths for a username",
			args:      args{"testuser", []string{"/some/path/to/%s/file"}},
			wantPaths: []string{"/some/path/to/testuser/file"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPaths := filePathsForCurrentUser(tt.args.username, tt.args.files); !reflect.DeepEqual(gotPaths, tt.wantPaths) {
				t.Errorf("filePathsForCurrentUser() = %v, want %v", gotPaths, tt.wantPaths)
			}
		})
	}
}

func Test_deleteFiles(t *testing.T) {
	os.Create("someTestFileToDelete.txt")
	type args struct {
		filePaths []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Can delete a list of files",
			args: args{[]string{"someTestFileToDelete.txt"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, path := range tt.args.filePaths {
				deleteFiles(tt.args.filePaths)
				if _, err := os.Stat(path); err == nil {
					t.Errorf("File exists %s", path)
				}
			}

		})
	}
}

func Test_copyToFolder(t *testing.T) {
	type args struct {
		source string
		target string
	}
	err := os.Mkdir("idea-copy", 0755)
	if err != nil {
		return
	}
	defer func() {
		err := os.RemoveAll("idea-copy")
		if err != nil {

		}
	}()
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Can copy a folder to a folder recursively",
			args: args{source: "./idea/misc.xml", target: "./idea-copy"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copyToFolder(tt.args.source, tt.args.target)
		})
	}
}
