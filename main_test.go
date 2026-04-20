package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Version
		wantErr bool
	}{
		{
			name:  "valid version",
			input: "1.2.3",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "zero version",
			input: "0.0.0",
			want:  Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:  "large numbers",
			input: "10.200.3000",
			want:  Version{Major: 10, Minor: 200, Patch: 3000},
		},
		{
			name:  "with whitespace",
			input: "  1.2.3  \n",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only whitespace",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "two parts",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "four parts",
			input:   "1.2.3.4",
			wantErr: true,
		},
		{
			name:    "non-numeric major",
			input:   "a.2.3",
			wantErr: true,
		},
		{
			name:    "non-numeric minor",
			input:   "1.b.3",
			wantErr: true,
		},
		{
			name:    "non-numeric patch",
			input:   "1.2.c",
			wantErr: true,
		},
		{
			name:    "negative major",
			input:   "-1.2.3",
			wantErr: true,
		},
		{
			name:    "with v prefix",
			input:   "v1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseVersion(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    string
	}{
		{
			name:    "zero",
			version: Version{0, 0, 0},
			want:    "0.0.0",
		},
		{
			name:    "typical",
			version: Version{1, 2, 3},
			want:    "1.2.3",
		},
		{
			name:    "large numbers",
			version: Version{10, 200, 3000},
			want:    "10.200.3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()
			if got != tt.want {
				t.Errorf("Version.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBumpPatch(t *testing.T) {
	tests := []struct {
		name  string
		input Version
		want  Version
	}{
		{
			name:  "from zero",
			input: Version{0, 0, 0},
			want:  Version{0, 0, 1},
		},
		{
			name:  "typical bump",
			input: Version{1, 2, 3},
			want:  Version{1, 2, 4},
		},
		{
			name:  "carry-over digits",
			input: Version{1, 2, 9},
			want:  Version{1, 2, 10},
		},
		{
			name:  "large patch",
			input: Version{1, 0, 999},
			want:  Version{1, 0, 1000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.BumpPatch()
			if got != tt.want {
				t.Errorf("BumpPatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadVersionFile(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "version")
		v, err := ReadVersionFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != (Version{0, 0, 0}) {
			t.Errorf("expected 0.0.0, got %s", v)
		}
	})

	t.Run("file with valid version", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "version")
		if err := os.WriteFile(path, []byte("2.5.10\n"), 0644); err != nil {
			t.Fatal(err)
		}
		v, err := ReadVersionFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != (Version{2, 5, 10}) {
			t.Errorf("expected 2.5.10, got %s", v)
		}
	})

	t.Run("file is empty", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "version")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		v, err := ReadVersionFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != (Version{0, 0, 0}) {
			t.Errorf("expected 0.0.0, got %s", v)
		}
	})

	t.Run("file with invalid content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "version")
		if err := os.WriteFile(path, []byte("not-a-version"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := ReadVersionFile(path)
		if err == nil {
			t.Fatal("expected error for invalid version content")
		}
	})
}

func TestWriteVersionFile(t *testing.T) {
	t.Run("write new file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "version")
		v := Version{1, 2, 3}
		if err := WriteVersionFile(path, v); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "1.2.3\n" {
			t.Errorf("expected %q, got %q", "1.2.3\n", string(data))
		}
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "version")
		if err := os.WriteFile(path, []byte("0.0.1\n"), 0644); err != nil {
			t.Fatal(err)
		}
		v := Version{0, 0, 2}
		if err := WriteVersionFile(path, v); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "0.0.2\n" {
			t.Errorf("expected %q, got %q", "0.0.2\n", string(data))
		}
	})
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "standard format",
			input: "scouratier/km_brain_fe",
			want:  "km_brain_fe",
		},
		{
			name:  "mixed case",
			input: "Scouratier/KM_Brain_FE",
			want:  "km_brain_fe",
		},
		{
			name:  "with whitespace",
			input: "  owner/repo  ",
			want:  "repo",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no slash",
			input:   "just-a-repo",
			wantErr: true,
		},
		{
			name:    "too many slashes",
			input:   "a/b/c",
			wantErr: true,
		},
		{
			name:    "empty repo name",
			input:   "owner/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractRepoName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractRepoName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ExtractRepoName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
